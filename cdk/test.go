package main

import (
	"runtime"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type TestStackProps struct {
	awscdk.StackProps
}

func NewPotStackStack(scope constructs.Construct, id string, props *TestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	_, filename, _, _ := runtime.Caller(1)
	appImage := awsecrassets.NewDockerImageAsset(stack, jsii.String("MyAsset"), &awsecrassets.DockerImageAssetProps{
		Directory: jsii.String("./../"),
	})

	taskDefinition := awsecs.NewFargateTaskDefinition(stack, jsii.String("MyTask"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(256),
		MemoryLimitMiB: jsii.Number(512),
	})

	container := taskDefinition.AddContainer(jsii.String("TaskDefinition"), &awsecs.ContainerDefinitionOptions{
		Image:     awsecs.ContainerImage_FromDockerImageAsset(appImage),
		Essential: jsii.Bool(true),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(8080),
				HostPort:      jsii.Number(80),
			},
		},
	})

	taskDefinition.SetDefaultContainer(container)

	cluster := awsecs.NewCluster(stack, jsii.String("EcsCluster"), &awsecs.ClusterProps{})
	awsecs.NewFargateService(stack, jsii.String("EcsService"), &awsecs.FargateServiceProps{
		Cluster:           cluster,
		TaskDefinition:    taskDefinition,
		DesiredCount:      jsii.Number(1),
		AssignPublicIp:    jsii.Bool(true),
		MaxHealthyPercent: jsii.Number(200),
		MinHealthyPercent: jsii.Number(0),

		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(60)),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewPotStackStack(app, "TestStack", &TestStackProps{
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
	return nil

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
}
