package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type BackendStackProps struct {
	*awscdk.StackProps
}

type BackendStackOutputs struct {
	ApiEndpoint awscdk.CfnOutput
	UserTable   awsdynamodb.Table
	JwtSecret   awssecretsmanager.Secret
}

func NewBackendStack(scope constructs.Construct, id string, props *BackendStackProps) (*awscdk.Stack, *BackendStackOutputs) {
	// func NewBackendStack(scope constructs.Construct, id string, props *BackendStackProps) *awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props.StackProps)

	// DynamoDB Table para Usuários
	userTable := awsdynamodb.NewTable(stack, jsii.String("Users"), &awsdynamodb.TableProps{
		TableName: jsii.String("kivaplus-usersv2"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("username"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		// BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		// RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // Para dev, use RETAIN em prod
	})

	// Add GSI para buscar por email
	userTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("email-index"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("email"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	// JWT Secret no Secrets Manager
	jwtSecret := awssecretsmanager.NewSecret(stack, jsii.String("JWTSecret"), &awssecretsmanager.SecretProps{
		SecretName:  jsii.String("kivaplus/jwt-secret"),
		Description: jsii.String("JWT Secret for authentication"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String(`{"jwt_secret": ""}`),
			GenerateStringKey:    jsii.String("jwt_secret"),
			ExcludeCharacters:    jsii.String(`"@/\`),
			PasswordLength:       jsii.Number(64),
		},
	})

	// IAM Role para Lambdas
	lambdaRole := awsiam.NewRole(stack, jsii.String("LambdaExecutionRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
		},
	})

	// Permissões para DynamoDB e Secrets Manager
	userTable.GrantReadWriteData(lambdaRole)
	jwtSecret.GrantRead(lambdaRole, nil)

	// Lambda Functions
	// awslambda.NewFunction(stack, jsii.String("HealthFunction"), &awslambda.FunctionProps{
	// 	Runtime: awslambda.Runtime_PROVIDED_AL2023(),
	// 	Code:    awslambda.AssetCode_FromAsset(jsii.String("backend/dist/health.zip"), nil),
	// 	Handler: jsii.String("main"),
	// 	Role:    lambdaRole,
	// 	Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
	// })

	authLambda := awslambda.NewFunction(stack, jsii.String("AuthFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("backend/dist/auth.zip"), nil),
		Handler: jsii.String("main"),
		Role:    lambdaRole,
		Environment: &map[string]*string{
			"USER_TABLE_NAME": userTable.TableName(),
			"JWT_SECRET_NAME": jwtSecret.SecretName(),
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
	})

	// userLambda := awslambda.NewFunction(stack, jsii.String("UserFunction"), &awslambda.FunctionProps{
	// 	Runtime: awslambda.Runtime_PROVIDED_AL2023(),
	// 	Code:    awslambda.Code_FromAsset(jsii.String("../backend/dist/user.zip"), nil),
	// 	Handler: jsii.String("bootstrap"),
	// 	Role:    lambdaRole,
	// 	Environment: &map[string]*string{
	// 		"USER_TABLE_NAME": userTable.TableName(),
	// 		"JWT_SECRET_NAME": jwtSecret.SecretName(),
	// 	},
	// 	Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
	// })

	// API Gateway
	api := awsapigateway.NewRestApi(stack, jsii.String("BackendApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("Kivaplus Backend API"),
		Description: jsii.String("API para autenticação e gestão de usuários"),
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowHeaders: jsii.Strings("Content-Type", "Authorization", "X-Amz-Date", "X-Api-Key", "X-Amz-Security-Token"),
			AllowMethods: jsii.Strings("GET", "POST", "DELETE", "PUT", "OPTIONS"),
			AllowOrigins: jsii.Strings("*"),
		},
		DeployOptions: &awsapigateway.StageOptions{
			LoggingLevel: awsapigateway.MethodLoggingLevel_INFO,
		},
		CloudWatchRole: jsii.Bool(true),
	})

	// Auth endpoints
	// authResource := api.Root().AddResource(jsii.String("auth"), nil)
	// authResource.AddMethod(jsii.String("POST"),
	// 	awsapigateway.NewLambdaIntegration(authLambda, &awsapigateway.LambdaIntegrationOptions{
	// 		Proxy: jsii.Bool(true),
	// 	}),
	// 	&awsapigateway.MethodOptions{})

	loginResource := api.Root().AddResource(jsii.String("login"), nil)
	loginResource.AddMethod(jsii.String("POST"),
		awsapigateway.NewLambdaIntegration(authLambda, &awsapigateway.LambdaIntegrationOptions{
			Proxy: jsii.Bool(true),
		}),
		&awsapigateway.MethodOptions{})

	registerResource := api.Root().AddResource(jsii.String("register"), nil)
	registerResource.AddMethod(jsii.String("POST"),
		awsapigateway.NewLambdaIntegration(authLambda, &awsapigateway.LambdaIntegrationOptions{
			Proxy: jsii.Bool(true),
		}),
		&awsapigateway.MethodOptions{})

	protectedResource := api.Root().AddResource(jsii.String("protected"), nil)
	protectedResource.AddMethod(jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(authLambda, &awsapigateway.LambdaIntegrationOptions{
			Proxy: jsii.Bool(true),
		}),
		&awsapigateway.MethodOptions{
			AuthorizationType: awsapigateway.AuthorizationType_NONE,
		})

	// User endpoints
	// userResource := api.Root().AddResource(jsii.String("user"), nil)
	// userResource.AddMethod(jsii.String("GET"),
	// 	awsapigateway.NewLambdaIntegration(userLambda, &awsapigateway.LambdaIntegrationOptions{
	// 		Proxy: jsii.Bool(true),
	// 	}),
	// 	&awsapigateway.MethodOptions{})

	// userResource.AddMethod(jsii.String("PUT"),
	// 	awsapigateway.NewLambdaIntegration(userLambda, &awsapigateway.LambdaIntegrationOptions{
	// 		Proxy: jsii.Bool(true),
	// 	}),
	// 	&awsapigateway.MethodOptions{})

	// Outputs
	apiEndpoint := awscdk.NewCfnOutput(stack, jsii.String("BackendApiEndpoint"), &awscdk.CfnOutputProps{
		Value:       api.Url(),
		Description: jsii.String("Backend API Gateway URL"),
		ExportName:  jsii.String("BackendApiEndpoint"),
	})

	outputs := &BackendStackOutputs{
		ApiEndpoint: apiEndpoint,
		UserTable:   userTable,
		JwtSecret:   jwtSecret,
	}

	return &stack, outputs
}
