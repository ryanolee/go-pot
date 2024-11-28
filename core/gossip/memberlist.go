package gossip

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/core/gossip/action"
	"github.com/ryanolee/ryan-pot/core/gossip/handler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	IMemberlist interface {
		Dispatch(*action.BroadcastAction)
		GetIpAddress() string
		Shutdown()
	}

	Memberlist struct {
		// The memberlist client
		client *memberlist.Memberlist

		// The delegate for memberlist events
		delegate *MessageDelegate

		// The channel to shutdown listeners for broadcast actions
		shutdownChan chan bool

		// The handler for broadcast actions. Is used to recieve events. Parse them and act upon them.
		handler handler.IBroadcastActionHandler

		// The number of connection attempts to make to other nodes in the cluster before giving up
		connectionAttempts int

		// The timeout for each connection attempt
		connectionTimeout time.Duration

		// The metadata for the current node
		nodeInfo *NodeInfo
	}

	// Metadata related to the current node
	NodeInfo struct {
		PeerIpAddresses []string
		IpAddress       string
	}
)

func NewMemberList(lf fx.Lifecycle, logger *zap.Logger, config *config.Config, broadcastHandler handler.IBroadcastActionHandler) (*Memberlist, error) {
	if !config.Cluster.Enabled {
		return nil, nil
	}

	//var err error
	var nodeInfo *NodeInfo
	var err error

	if config.Cluster.Mode == "fargate_ecs" {
		nodeInfo, err = gatherFargateNodeInfo()
		if err != nil {
			return nil, err
		}
	} else if config.Cluster.Mode == "lan" || config.Cluster.Mode == "wan" {
		nodeInfo = &NodeInfo{
			IpAddress:       config.Cluster.AdvertiseIp,
			PeerIpAddresses: config.Cluster.KnownPeerIps,
		}
	} else {
		return nil, errors.New("No valid cluster mode specified. Must be one of: fargate_ecs, lan, wan")
	}

	// CFG initial setup
	cfg := memberlist.DefaultLANConfig()
	cfg.BindPort = config.Cluster.BindPort
	cfg.BindAddr = nodeInfo.IpAddress

	// Bind Event delegate
	evtDelegate := NewMessageEventDelegate()
	cfg.Events = evtDelegate

	// Bind Broadcast delegate
	msgDelegate := NewMessageDelegate(evtDelegate)
	cfg.Delegate = msgDelegate

	// Black hole logger
	if !config.Cluster.EnableLogging {
		nullLogger := log.New(io.Discard, "", 0)
		cfg.Logger = nullLogger
	}

	client, err := memberlist.Create(cfg)
	if err != nil {
		return nil, err
	}

	memberList := &Memberlist{
		client:             client,
		delegate:           msgDelegate,
		handler:            broadcastHandler,
		shutdownChan:       make(chan bool),
		connectionAttempts: config.Cluster.ConnectionAttempts,
		connectionTimeout:  time.Duration(config.Cluster.ConnectionTimeout) * time.Second,
		nodeInfo:           nodeInfo,
	}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := memberList.Join(); err != nil {
					zap.L().Sugar().Error("Failed to join member list", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			memberList.Shutdown()
			return nil
		},
	})

	return memberList, nil
}

func (m *Memberlist) Join() error {
	for i := 0; i < m.connectionAttempts; i++ {
		nodes, err := m.client.Join(m.nodeInfo.PeerIpAddresses)
		if err != nil && nodes == 0 {
			zap.L().Sugar().Warnw("Failed to connect to other nodes", "err", err, "attempt", i)
			time.Sleep(m.connectionTimeout)
			continue
		}
		break
	}

	if m.client.NumMembers() < 2 {
		return errors.New("failed to connect to any other nodes in the cluster")
	}

	zap.L().Sugar().Infow("Connected to other nodes", "peers", m.client.NumMembers(), "ip", m.GetIpAddress())
	m.listenForBroadcastActions()

	return nil
}

// GetIpAddress returns the IP address of the current node
func (m *Memberlist) GetIpAddress() string {
	return m.client.LocalNode().Addr.String()
}

// Broadcasts a message to peer nodes in the cluster
func (m *Memberlist) Dispatch(broadcast *action.BroadcastAction) {
	zap.L().Sugar().Infow("Broadcasting action", "action", broadcast.Action, "data", broadcast.Data)
	m.delegate.Broadcasts.QueueBroadcast(broadcast)
}

// ListenForBroadcastActions listens for broadcast actions from other nodes in the cluster
func (m *Memberlist) listenForBroadcastActions() {
	go func() {
		for {
			select {
			case msg := <-m.delegate.MessageChan:
				action, err := action.ParseBroadcastAction(msg)
				if err != nil {
					zap.L().Sugar().Warnw("Failed to parse broadcast action", "err", err, "data", string(msg))
					continue
				}
				go m.handler.Handle(action)
			case <-m.shutdownChan:
				return
			}
		}
	}()
}

// nolint:errcheck
func (m *Memberlist) Shutdown() {
	m.client.Shutdown()
	m.shutdownChan <- true
}
