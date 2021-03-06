package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

type GoBitcoinStackProps struct {
	awscdk.StackProps
}

func NewGoBitcoinStack(scope constructs.Construct, id string, props *GoBitcoinStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// -------------------------
	// Secret Manager
	// -------------------------
	secret := awssecretsmanager.NewSecret(stack, jsii.String("bitcoin-secret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("buy-bitcoin-secret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			GenerateStringKey:    jsii.String("key"),
			SecretStringTemplate: jsii.String(`{ "key": "", "secret": "" }`),
		},
	})

	// -------------------------
	// IAM
	// -------------------------
	policyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("secretsmanager:GetResourcePolicy"),
			jsii.String("secretsmanager:GetSecretValue"),
		},
		Resources: &[]*string{secret.ArnForPolicies()},
	})
	role := awsiam.NewRole(stack, jsii.String("lambda-role"), &awsiam.RoleProps{
		RoleName:  jsii.String("go-bitcoin-lambda-role"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
		},
	})
	role.AddToPolicy(policyStatement)

	// -------------------------
	// Lambda
	// -------------------------
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	awslambda.NewFunction(stack, jsii.String("bitcoin"), &awslambda.FunctionProps{
		FunctionName: jsii.String("bitcoin-lambda"),
		Code:         awslambda.Code_FromAsset(jsii.String(filepath.Join(curDir, "/lambda/function.zip")), nil),
		Handler:      jsii.String("main"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		Role:         role,
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewGoBitcoinStack(app, "GoBitcoinStack", &GoBitcoinStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	// ---------------------------------------------------------------------------
	return &awscdk.Environment{
		Region: jsii.String("ap-northeast-1"),
	}

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
