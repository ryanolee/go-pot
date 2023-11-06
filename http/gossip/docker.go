package gossip

import (
	"os"
	"strings"
)

func inDockerContainer() bool {
	return os.Getenv("DOCKER_KNOWN_PEERS") != ""
}

func gatherDockerNodeInfo() *NodeInfo {
	return &NodeInfo{
		IpAddress:       os.Getenv("DOCKER_IP_ADDRESS"),
		PeerIpAddresses: strings.Split(os.Getenv("DOCKER_KNOWN_PEERS"), ","),
	}
}
