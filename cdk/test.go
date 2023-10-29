package main

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"

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

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to obtain filepath")
	}

	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("172.31.0.0/24")),
		MaxAzs: jsii.Number(2),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				CidrMask: jsii.Number(26),
				Name:     jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
	})

	appImage := awsecrassets.NewDockerImageAsset(stack, jsii.String("MyAsset"), &awsecrassets.DockerImageAssetProps{
		Directory: jsii.String(path.Join(filepath.Dir(filename), "..")),
		Exclude: jsii.Strings("./cdk"),
	})

	taskDefinition := awsecs.NewFargateTaskDefinition(stack, jsii.String("MyTask"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(256),
		MemoryLimitMiB: jsii.Number(512),
		RuntimePlatform: &awsecs.RuntimePlatform{
			CpuArchitecture: awsecs.CpuArchitecture_X86_64(),
			OperatingSystemFamily: awsecs.OperatingSystemFamily_LINUX(),
		},
	})

	logGroup := awslogs.NewLogGroup(stack, jsii.String("LogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/ryan-pot/nodes"),
		Retention:    awslogs.RetentionDays_ONE_WEEK,
	})

	container := taskDefinition.AddContainer(jsii.String("TaskDefinition"), &awsecs.ContainerDefinitionOptions{
		Image:     awsecs.ContainerImage_FromDockerImageAsset(appImage),
		Essential: jsii.Bool(true),
		Logging: awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
			LogGroup: logGroup,
			StreamPrefix: jsii.String("/ryan-pot"),
		}),
		
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(8080),
			},
		},
	})

	taskDefinition.SetDefaultContainer(container)

	cluster := awsecs.NewCluster(stack, jsii.String("EcsCluster"), &awsecs.ClusterProps{
		Vpc: vpc,
	})
	awsecs.NewFargateService(stack, jsii.String("EcsService"), &awsecs.FargateServiceProps{
		Cluster:           cluster,
		CapacityProviderStrategies: &[]*awsecs.CapacityProviderStrategy{
			{
				CapacityProvider: jsii.String("FARGATE_SPOT"),
				Weight:           jsii.Number(1),
			},
		},
		VpcSubnets: 	  &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		TaskDefinition:    taskDefinition,
		DesiredCount:      jsii.Number(1),
		AssignPublicIp:    jsii.Bool(true),
		MaxHealthyPercent: jsii.Number(200),
		MinHealthyPercent: jsii.Number(0),
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
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String("eu-west-1"),
	   }
}
