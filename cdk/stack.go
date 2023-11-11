package main

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	metricsRegion   = "eu-west-1"
	nodesPerCluster = 48
)

var (
	potRegions = []string{"eu-west-1", "us-east-1", "ap-northeast-1"}
)

type (
	MetricsServerCreds struct {
		Username *string
		Password *string
	}
	MetricsStackProps struct {
		StackProps         awscdk.StackProps
		MetricsServerCreds *MetricsServerCreds
	}
	MetricsStack struct {
		Stack         awscdk.Stack
		MetricsServer awsec2.Instance
	}

	PotStackProps struct {
		StackProps         awscdk.StackProps
		MetricsServerCreds *MetricsServerCreds
		MetricsServer      awsec2.Instance
		NodeCount          int
	}
)

func NewMetricsStack(scope constructs.Construct, id string, props *MetricsStackProps) MetricsStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	sprops.CrossRegionReferences = jsii.Bool(true)
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("172.31.1.0/24")),
		MaxAzs:      jsii.Number(1),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				CidrMask:   jsii.Number(26),
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
	})

	// EC2 Prometheus Push Gateway
	pushGatewaySg := awsec2.NewSecurityGroup(stack, jsii.String("PushGatewaySecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc: vpc,
	})
	pushGatewaySg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(9092)), jsii.String("Ingress from prometheus (Internet)"), jsii.Bool(false))
	pushGatewaySg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(9093)), jsii.String("Ingress from prometheus (Internet)"), jsii.Bool(false))

	pushGateway := awsec2.NewInstance(stack, jsii.String("PrometheusMetricsNode"), &awsec2.InstanceProps{
		InstanceType: awsec2.NewInstanceType(jsii.String("t3.small")),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		SecurityGroup: pushGatewaySg,
	})

	//ssmData := awsssm.StringParameter_FromStringParameterName(stack, jsii.String("SsmGrafanaCloud"), jsii.String("/ryan-pot/grafana-cloud-key"))
	//ssmData.GrantRead(pushGateway.Role())
	pushGateway.UserData().AddCommands(
		// Install SSM Agent
		jsii.String("sudo yum install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"),
		jsii.String("sudo systemctl enable amazon-ssm-agent"),
		jsii.String("sudo systemctl start amazon-ssm-agent"),

		// Install utils
		jsii.String("sudo yum install -y gettext envsubst"),

		// Setup Prometheus Push Gateway
		jsii.String("sudo useradd -M -r -s /bin/false pushgateway"),
		jsii.String("wget https://github.com/prometheus/pushgateway/releases/download/v1.2.0/pushgateway-1.2.0.linux-amd64.tar.gz"),
		jsii.String("tar xvfz pushgateway-1.2.0.linux-amd64.tar.gz"),
		jsii.String("sudo cp pushgateway-1.2.0.linux-amd64/pushgateway /usr/local/bin/"),
		jsii.String("sudo chown pushgateway:pushgateway /usr/local/bin/pushgateway"),
		jsii.String(`echo "[Unit]
Description=Prometheus Pushgateway
Wants=network-online.target
After=network-online.target

[Service]
User=pushgateway
Group=pushgateway
Type=simple
ExecStart=/usr/local/bin/pushgateway
[Install]
WantedBy=multi-user.target" > /etc/systemd/system/pushgateway.service`),
		jsii.String("sudo systemctl enable pushgateway"),
		jsii.String("sudo systemctl start pushgateway"),

		// Install prometheus
		jsii.String("sudo useradd --no-create-home --shell /bin/false prometheus"),
		jsii.String("sudo mkdir /etc/prometheus /var/lib/prometheus"),
		jsii.String("sudo chown prometheus:prometheus /etc/prometheus /var/lib/prometheus"),
		jsii.String("cd ~"),
		jsii.String("curl -LO https://github.com/prometheus/prometheus/releases/download/v2.45.1/prometheus-2.45.1.linux-amd64.tar.gz"),
		jsii.String("tar -xvf prometheus-2.45.1.linux-amd64.tar.gz"),
		jsii.String("sudo cp -p ./prometheus-2.45.1.linux-amd64/prometheus /usr/local/bin"),
		jsii.String("sudo chown prometheus:prometheus /usr/local/bin/prom*"),
		jsii.String("sudo cp -r ./prometheus-2.45.1.linux-amd64/consoles /etc/prometheus"),
		jsii.String("sudo cp -r ./prometheus-2.45.1.linux-amd64/console_libraries /etc/prometheus"),
		jsii.String("sudo chown -R prometheus:prometheus /etc/prometheus/consoles /etc/prometheus/console_libraries"),
		jsii.String(`echo "global:
  scrape_interval: 1m
  evaluation_interval: 1m
  scrape_timeout: 2s
scrape_configs:
- job_name: push_gateway
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets: ['localhost:9091']
    labels:
      service: 'prom-pushgateway'
" > /etc/prometheus/prometheus.yml`),
		// Pull config from SSM
		//awscdk.Fn_Sub(jsii.String("aws ssm get-parameter --region ${REGION} --name ${NAME} --with-decryption --query Parameter.Value --output text >> /etc/prometheus/prometheus.yml"), &map[string]*string{
		//	"REGION": props.Env.Region,
		//	"NAME":   ssmData.ParameterName(),
		//}),

		jsii.String(`echo "[Unit]
Description=PromServer
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
--config.file /etc/prometheus/prometheus.yml \
--storage.tsdb.path /var/lib/prometheus/ \
--web.console.templates=/etc/prometheus/consoles \
--web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/prometheus.service`),
		jsii.String("sudo systemctl daemon-reload"),
		jsii.String("sudo systemctl enable prometheus"),
		jsii.String("sudo systemctl start prometheus"),

		// Setup and install nginx
		jsii.String("sudo amazon-linux-extras install nginx1 -y"),
		jsii.String("sudo chkconfig nginx on"),
		jsii.String("sudo service nginx start"),
		jsii.String(`sudo echo "server {
	listen *:9092;
		location / {
			auth_basic             "Restricted";
			auth_basic_user_file   .htpasswd;

			proxy_pass              http://localhost:9090;
		}
	}
	server {
	listen *:9093;
		location / {
			auth_basic             "Restricted";
			auth_basic_user_file   .htpasswd;

			proxy_pass              http://localhost:9091;
		}
	}" > /etc/nginx/conf.d/pushgateway.conf`),
		jsii.String(`sudo yum install httpd-tools -y`),
		awscdk.Fn_Sub(jsii.String("sudo htpasswd -c -b /etc/nginx/.htpasswd ${USERNAME} ${PASSWORD}"), &map[string]*string{
			"USERNAME": props.MetricsServerCreds.Username,
			// @todo Pull this from Secrets Manager
			"PASSWORD": props.MetricsServerCreds.Password,
		}),
		jsii.String("sudo service nginx restart"),
	)

	pushGateway.Role().AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")))
	return MetricsStack{
		Stack:         stack,
		MetricsServer: pushGateway,
	}
}

func NewPotStackStack(scope constructs.Construct, id string, props *PotStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	sprops.CrossRegionReferences = jsii.Bool(true)
	stack := awscdk.NewStack(scope, &id, &sprops)

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to obtain filepath")
	}

	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("172.31.0.0/24")),
		MaxAzs:      jsii.Number(1),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				CidrMask:   jsii.Number(26),
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
	})

	// EC2 Prometheus Push Gateway
	pushGatewaySg := awsec2.NewSecurityGroup(stack, jsii.String("PushGatewaySecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc: vpc,
	})
	pushGatewaySg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(9092)), jsii.String("Ingress from prometheus (Internet)"), jsii.Bool(false))
	pushGatewaySg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(9093)), jsii.String("Ingress from prometheus (Internet)"), jsii.Bool(false))

	pushGateway := awsec2.NewInstance(stack, jsii.String("PrometheusMetricsNode"), &awsec2.InstanceProps{
		InstanceType: awsec2.NewInstanceType(jsii.String("t3.micro")),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		SecurityGroup: pushGatewaySg,
	})

	//ssmData := awsssm.StringParameter_FromStringParameterName(stack, jsii.String("SsmGrafanaCloud"), jsii.String("/ryan-pot/grafana-cloud-key"))
	//ssmData.GrantRead(pushGateway.Role())
	pushGateway.UserData().AddCommands(
		// Install SSM Agent
		jsii.String("sudo yum install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"),
		jsii.String("sudo systemctl enable amazon-ssm-agent"),
		jsii.String("sudo systemctl start amazon-ssm-agent"),

		// Install utils
		jsii.String("sudo yum install -y gettext envsubst"),

		// Setup Prometheus Push Gateway
		jsii.String("sudo useradd -M -r -s /bin/false pushgateway"),
		jsii.String("wget https://github.com/prometheus/pushgateway/releases/download/v1.2.0/pushgateway-1.2.0.linux-amd64.tar.gz"),
		jsii.String("tar xvfz pushgateway-1.2.0.linux-amd64.tar.gz"),
		jsii.String("sudo cp pushgateway-1.2.0.linux-amd64/pushgateway /usr/local/bin/"),
		jsii.String("sudo chown pushgateway:pushgateway /usr/local/bin/pushgateway"),
		jsii.String(`echo "[Unit]
Description=Prometheus Pushgateway
Wants=network-online.target
After=network-online.target

[Service]
User=pushgateway
Group=pushgateway
Type=simple
ExecStart=/usr/local/bin/pushgateway
[Install]
WantedBy=multi-user.target" > /etc/systemd/system/pushgateway.service`),
		jsii.String("sudo systemctl enable pushgateway"),
		jsii.String("sudo systemctl start pushgateway"),

		// Install prometheus
		jsii.String("sudo useradd --no-create-home --shell /bin/false prometheus"),
		jsii.String("sudo mkdir /etc/prometheus /var/lib/prometheus"),
		jsii.String("sudo chown prometheus:prometheus /etc/prometheus /var/lib/prometheus"),
		jsii.String("cd ~"),
		jsii.String("curl -LO https://github.com/prometheus/prometheus/releases/download/v2.45.1/prometheus-2.45.1.linux-amd64.tar.gz"),
		jsii.String("tar -xvf prometheus-2.45.1.linux-amd64.tar.gz"),
		jsii.String("sudo cp -p ./prometheus-2.45.1.linux-amd64/prometheus /usr/local/bin"),
		jsii.String("sudo chown prometheus:prometheus /usr/local/bin/prom*"),
		jsii.String("sudo cp -r ./prometheus-2.45.1.linux-amd64/consoles /etc/prometheus"),
		jsii.String("sudo cp -r ./prometheus-2.45.1.linux-amd64/console_libraries /etc/prometheus"),
		jsii.String("sudo chown -R prometheus:prometheus /etc/prometheus/consoles /etc/prometheus/console_libraries"),
		jsii.String(`echo "global:
  scrape_interval: 1m
  evaluation_interval: 1m
  scrape_timeout: 2s
scrape_configs:
- job_name: push_gateway
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets: ['localhost:9091']
    labels:
      service: 'prom-pushgateway'
" > /etc/prometheus/prometheus.yml`),
		// Pull config from SSM
		//awscdk.Fn_Sub(jsii.String("aws ssm get-parameter --region ${REGION} --name ${NAME} --with-decryption --query Parameter.Value --output text >> /etc/prometheus/prometheus.yml"), &map[string]*string{
		//	"REGION": props.Env.Region,
		//	"NAME":   ssmData.ParameterName(),
		//}),

		jsii.String(`echo "[Unit]
Description=PromServer
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
--config.file /etc/prometheus/prometheus.yml \
--storage.tsdb.path /var/lib/prometheus/ \
--web.console.templates=/etc/prometheus/consoles \
--web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/prometheus.service`),
		jsii.String("sudo systemctl daemon-reload"),
		jsii.String("sudo systemctl enable prometheus"),
		jsii.String("sudo systemctl start prometheus"),

		// Setup and install nginx
		jsii.String("sudo amazon-linux-extras install nginx1 -y"),
		jsii.String("sudo chkconfig nginx on"),
		jsii.String("sudo service nginx start"),
		jsii.String(`sudo echo "server {
	listen *:9092;
		location / {
			auth_basic             "Restricted";
			auth_basic_user_file   .htpasswd;

			proxy_pass              http://localhost:9090;
		}
	}
	server {
	listen *:9093;
		location / {
			auth_basic             "Restricted";
			auth_basic_user_file   .htpasswd;

			proxy_pass              http://localhost:9091;
		}
	}" > /etc/nginx/conf.d/pushgateway.conf`),
		jsii.String(`sudo yum install httpd-tools -y`),
		awscdk.Fn_Sub(jsii.String("sudo htpasswd -c -b /etc/nginx/.htpasswd ${USERNAME} ${PASSWORD}"), &map[string]*string{
			"USERNAME": jsii.String("ryan-pot"),
			// @todo Pull this from Secrets Manager
			"PASSWORD": jsii.String("SOME_SECRET"),
		}),
		jsii.String("sudo service nginx restart"),
	)

	pushGateway.Role().AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")))

	appImage := awsecrassets.NewDockerImageAsset(stack, jsii.String("EcrAsset"), &awsecrassets.DockerImageAssetProps{
		Directory: jsii.String(path.Join(filepath.Dir(filename), "..")),
		Exclude:   jsii.Strings("./cdk"),
		Target:    jsii.String("prod"),
	})

	taskDefinition := awsecs.NewFargateTaskDefinition(stack, jsii.String("TaskDefinition"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(256),
		MemoryLimitMiB: jsii.Number(512),
		RuntimePlatform: &awsecs.RuntimePlatform{
			CpuArchitecture:       awsecs.CpuArchitecture_X86_64(),
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
			LogGroup:     logGroup,
			StreamPrefix: jsii.String("/ryan-pot"),
		}),

		Environment: &map[string]*string{
			"PUSH_GATEWAY_ADDRESS": awscdk.Fn_Sub(jsii.String("${PUBLIC_DNS}:9093"), &map[string]*string{
				"PUBLIC_DNS": props.MetricsServer.InstancePublicDnsName(),
			}),
			"PUSH_GATEWAY_USERNAME": props.MetricsServerCreds.Username,
			"PUSH_GATEWAY_PASSWORD": props.MetricsServerCreds.Password,
			"PUSH_GATEWAY_REGION":   stack.Region(),
		},

		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(80),
			},
			{
				ContainerPort: jsii.Number(7947),
			},
		},
	})

	taskDefinition.SetDefaultContainer(container)

	cluster := awsecs.NewCluster(stack, jsii.String("EcsCluster"), &awsecs.ClusterProps{
		Vpc:               vpc,
		ContainerInsights: jsii.Bool(false),
	})

	serviceSg := awsec2.NewSecurityGroup(stack, jsii.String("SecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc: vpc,
	})

	pushGatewaySg.AddIngressRule(serviceSg, awsec2.Port_Tcp(jsii.Number(9091)), jsii.String("Allow Prometheus Push Gateway traffic"), jsii.Bool(false))

	serviceSg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("Allow HTTP traffic from anywhere"), jsii.Bool(false))
	serviceSg.AddIngressRule(awsec2.Peer_Ipv4(jsii.String("172.31.0.0/24")), awsec2.Port_AllTraffic(), jsii.String("Allow internal traffic"), jsii.Bool(false))
	awsecs.NewFargateService(stack, jsii.String("EcsService"), &awsecs.FargateServiceProps{
		Cluster: cluster,
		CapacityProviderStrategies: &[]*awsecs.CapacityProviderStrategy{
			{
				CapacityProvider: jsii.String("FARGATE_SPOT"),
				Weight:           jsii.Number(1),
			},
		},
		VpcSubnets:        &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		TaskDefinition:    taskDefinition,
		DesiredCount:      jsii.Number(props.NodeCount),
		AssignPublicIp:    jsii.Bool(true),
		MaxHealthyPercent: jsii.Number(200),
		MinHealthyPercent: jsii.Number(0),
		SecurityGroups:    &[]awsec2.ISecurityGroup{serviceSg},
	})

	taskDefinitionPolicy := awsiam.NewPolicy(stack, jsii.String("TaskDefinitionPolicy"), &awsiam.PolicyProps{
		Statements: &[]awsiam.PolicyStatement{awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Actions: &[]*string{
				jsii.String("ecs:ListTasks"),
				jsii.String("ecs:DescribeTasks"),
			},
			Effect: awsiam.Effect_ALLOW,
			Resources: &[]*string{
				jsii.String("*"),
			},
		})},
	})
	taskDefinition.TaskRole().AttachInlinePolicy(taskDefinitionPolicy)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// @todo - Move this to secrets manager at some stage
	metricsServerPassword, ok := app.Node().TryGetContext(jsii.String("MetricsServerPassword")).(string)

	if !ok || metricsServerPassword == "" {
		fmt.Println("Watning! MetricsServerPassword must be set  in context (--context \"MetricsServerPassword ...\")")
		metricsServerPassword = "foobar"
	}

	metricsServerCreds := &MetricsServerCreds{
		Username: jsii.String("go-pot"),
		Password: jsii.String(string(metricsServerPassword)),
	}

	metricsStack := NewMetricsStack(app, "GoPotMetricsStack", &MetricsStackProps{
		StackProps: awscdk.StackProps{
			Env: env(metricsRegion),
		},
		MetricsServerCreds: metricsServerCreds,
	})

	for _, region := range potRegions {
		NewPotStackStack(app, fmt.Sprintf("GoPotStack-%s", region), &PotStackProps{
			StackProps: awscdk.StackProps{
				Env: env(region),
			},
			MetricsServerCreds: metricsServerCreds,
			MetricsServer:      metricsStack.MetricsServer,
			NodeCount:          nodesPerCluster,
		})
	}

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env(region string) *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("849652302708"),
		Region:  jsii.String(region),
	}
}
