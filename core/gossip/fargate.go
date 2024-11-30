package gossip

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"
	"github.com/thoas/go-funk"
)

const (
	privateIpPrefix = "172.31."
)

func gatherFargateNodeInfo() (*NodeInfo, error) {
	meta, err := metadata.GetContainerV4(context.Background(), &http.Client{})

	if err != nil {
		return nil, err
	}

	if len(meta.Networks) == 0 {
		return nil, errors.New("metadata does not contain network information")
	}

	privateIpAddresses := funk.Filter(meta.Networks[0].IPv4Addresses, func(ipAddress string) bool {
		return strings.HasPrefix(ipAddress, privateIpPrefix)
	}).([]string)

	if len(privateIpAddresses) == 0 {
		return nil, errors.New("metadata does not contain private IPv4 address")
	}

	taskIpAddress := privateIpAddresses[0]

	// Fetch Peer IP Addresses of ecs tasks
	var peerAddresses []string
	if peerAddresses, err = getEcsIpAddresses(meta.Labels.EcsCluster); err != nil {
		return nil, err
	}

	// Filter out the current task's IP address
	peerAddresses = funk.Filter(peerAddresses, func(ipAddress string) bool {
		return ipAddress != taskIpAddress
	}).([]string)

	return &NodeInfo{
		IpAddress:       "0.0.0.0",
		PeerIpAddresses: peerAddresses,
	}, nil
}

func getEcsIpAddresses(clusterName string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	// Create ECS client
	client := ecs.NewFromConfig(cfg)
	listTaskOutput, err := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster: &clusterName,
	})

	if err != nil {
		return nil, err
	}

	describeTaskOutput, err := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Cluster: &clusterName,
		Tasks:   listTaskOutput.TaskArns,
	})

	if err != nil {
		return nil, err
	}

	ipAddresses := []string{}
	for _, task := range describeTaskOutput.Tasks {
		if len(task.Containers) == 0 {
			continue
		}

		container := task.Containers[0]

		if len(container.NetworkInterfaces) == 0 {
			continue
		}

		network := container.NetworkInterfaces[0]
		ipAddresses = append(ipAddresses, *network.PrivateIpv4Address)
	}

	return ipAddresses, nil
}
