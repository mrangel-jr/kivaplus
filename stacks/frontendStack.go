package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsamplify"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type FrontendStackProps struct {
	*awscdk.StackProps
	BackendOutputs *BackendStackOutputs
}

func NewFrontendStack(scope constructs.Construct, id string, props *FrontendStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props.StackProps)

	apiEndpoint := props.BackendOutputs.ApiEndpoint.Value().(string)

	// Amplify App
	amplifyApp := awsamplify.NewCfnApp(stack, jsii.String("FrontendApp"), &awsamplify.CfnAppProps{
		Name:        jsii.String("kivaplus-frontend"),
		Description: jsii.String("Kivaplus Frontend Next.js App"),
		Repository:  jsii.String("https://github.com/mrangel-jr/kivaplus"),
		AccessToken: jsii.String("{{resolve:secretsmanager:github-token:SecretString:token}}"),
		EnvironmentVariables: &[]*awsamplify.CfnApp_EnvironmentVariableProperty{
			{
				Name:  jsii.String("NEXT_PUBLIC_API_URL"),
				Value: &apiEndpoint,
			},
			{
				Name:  jsii.String("AMPLIFY_MONOREPO_APP_ROOT"),
				Value: jsii.String("frontend"),
			},
		},
		BuildSpec: jsii.String(`version: 1
applications:
  - appRoot: frontend
    frontend:
      phases:
        preBuild:
          commands:
            - npm install
        build:
          commands:
            - npm run build
      artifacts:
        baseDirectory: .next
        files:
          - '**/*'
      cache:
        paths:
          - node_modules/**/*`),
	})

	// Branch principal (USANDO a variável!)
	mainBranch := awsamplify.NewCfnBranch(stack, jsii.String("MainBranch"), &awsamplify.CfnBranchProps{
		AppId:           amplifyApp.AttrAppId(),
		BranchName:      jsii.String("main"),
		Stage:           jsii.String("PRODUCTION"),
		EnableAutoBuild: jsii.Bool(true),
		EnvironmentVariables: &[]*awsamplify.CfnBranch_EnvironmentVariableProperty{
			{
				Name:  jsii.String("NEXT_PUBLIC_API_URL"),
				Value: &apiEndpoint,
			},
			{
				Name:  jsii.String("_LIVE_UPDATES"),
				Value: jsii.String(`[{"name":"Next.js version","pkg":"next","type":"npm","version":"latest"}]`),
			},
		},
	})

	// Branch de desenvolvimento
	devBranch := awsamplify.NewCfnBranch(stack, jsii.String("DevBranch"), &awsamplify.CfnBranchProps{
		AppId:           amplifyApp.AttrAppId(),
		BranchName:      jsii.String("develop"),
		Stage:           jsii.String("DEVELOPMENT"),
		EnableAutoBuild: jsii.Bool(true),
		EnvironmentVariables: &[]*awsamplify.CfnBranch_EnvironmentVariableProperty{
			{
				Name:  jsii.String("NEXT_PUBLIC_API_URL"),
				Value: &apiEndpoint,
			},
		},
	})

	// Outputs usando as URLs automáticas do Amplify
	awscdk.NewCfnOutput(stack, jsii.String("AmplifyAppId"), &awscdk.CfnOutputProps{
		Value:       amplifyApp.AttrAppId(),
		Description: jsii.String("Amplify App ID"),
	})

	// URL da branch main (PRODUÇÃO)
	awscdk.NewCfnOutput(stack, jsii.String("FrontendMainUrl"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Sub(jsii.String("https://${BranchName}.${AppId}.amplifyapp.com"), &map[string]*string{
			"BranchName": mainBranch.AttrBranchName(),
			"AppId":      amplifyApp.AttrAppId(),
		}),
		Description: jsii.String("Frontend Main Branch URL (Production)"),
		ExportName:  jsii.String("FrontendMainUrl"),
	})

	// URL da branch develop (DESENVOLVIMENTO)
	awscdk.NewCfnOutput(stack, jsii.String("FrontendDevUrl"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Sub(jsii.String("https://${BranchName}.${AppId}.amplifyapp.com"), &map[string]*string{
			"BranchName": devBranch.AttrBranchName(),
			"AppId":      amplifyApp.AttrAppId(),
		}),
		Description: jsii.String("Frontend Dev Branch URL (Development)"),
		ExportName:  jsii.String("FrontendDevUrl"),
	})

	// URL genérica da app (aponta para main por padrão)
	awscdk.NewCfnOutput(stack, jsii.String("FrontendUrl"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Sub(jsii.String("https://${AppId}.amplifyapp.com"), &map[string]*string{
			"AppId": amplifyApp.AttrAppId(),
		}),
		Description: jsii.String("Frontend App URL (redirects to main)"),
	})

	return stack
}
