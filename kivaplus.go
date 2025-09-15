package main

import (
	"kivaplus/stacks"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)

	// Backend Stack (deve ser deployed primeiro)
	stacks.NewBackendStack(app, "KivaPlusBackendStack", &stacks.BackendStackProps{
		StackProps: &awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Backend API Stack com JWT e Secrets Manager"),
		},
	})

	// Frontend Stack (depende do Backend)
	// frontendStack := stacks.NewFrontendStack(app, "KivaPlusFrontendStack", &stacks.FrontendStackProps{
	// 	StackProps: &awscdk.StackProps{
	// 		Env:         env(),
	// 		Description: jsii.String("Frontend Next.js Stack"),
	// 	},
	// 	BackendOutputs: backendOutputs,
	// })

	// frontendStack.AddDependency(*backendStack, jsii.String("Frontend depends on Backend"))

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
	// return &awscdk.Environment{
	// 	Account: jsii.String("000000000000"),
	// 	Region:  jsii.String("us-east-1"),
	// }
}

// Uncomment if you know exactly what account and region you want to deploy
// the stack to. This is the recommendation for production stacks.
//---------------------------------------------------------------------------
// return &awscdk.Environment{
//  Account: jsii.String("123456789012"),
//  Region:  jsii.String("us-east-1"),
// }

// Uncomment to specialize this stack for the AWS Account and Region that are
// implied by the current CLI configuration. This is recommended for dev
// stacks.
//---------------------------------------------------------------------------
// return &awscdk.Environment{
//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
// }
