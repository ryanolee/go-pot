package gossip

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"go.uber.org/zap"
)

//"github.com/hashicorp/memberlist"

const connectionAttempts = 5
const connectionTimeout = time.Second * 2

type (
	MemberlistOpts struct {
		OnAction     func(*action.BroadcastAction)
		SuppressLogs bool
	}

	Memberlist struct {
		client       *memberlist.Memberlist
		delegate     *MessageDelegate
		shutdownChan chan bool
		onAction     func(*action.BroadcastAction)
	}

	NodeInfo struct {
		PeerIpAddresses []string
		IpAddress       string
	}
)

func NewMemberList(opts *MemberlistOpts) (*Memberlist, error) {
	//var err error
	var nodeInfo *NodeInfo
	var err error

	if inFargateCluster() {
		nodeInfo, err = gatherFargateNodeInfo()
		if err != nil {
			return nil, err
		}
	} else if inDockerContainer() {
		nodeInfo = gatherDockerNodeInfo()
	}

	// CFG initial setup
	cfg := memberlist.DefaultLANConfig()
	cfg.BindPort = 7947
	cfg.BindAddr = nodeInfo.IpAddress

	// Bind Event delegate
	evtDelegate := NewMessageEventDelegate()
	cfg.Events = evtDelegate

	// Bind Broadcast delegate
	msgDelegate := NewMessageDelegate(evtDelegate)
	cfg.Delegate = msgDelegate

	// Black hole logger
	if opts.SuppressLogs {
		nullLogger := log.New(io.Discard, "", 0)
		cfg.Logger = nullLogger
	}

	client, err := memberlist.Create(cfg)
	if err != nil {
		return nil, err
	}

	for i := 0; i < connectionAttempts; i++ {
		nodes, err := client.Join(nodeInfo.PeerIpAddresses)
		if err != nil && nodes == 0 {
			zap.L().Sugar().Warnw("Failed to connect to other nodes", "err", err, "attempt", i)
			time.Sleep(connectionTimeout)
			continue
		}
		break
	}

	if client.NumMembers() < 2 {
		return nil, errors.New("failed to connect to other nodes")
	}

	if opts.OnAction == nil {
		opts.OnAction = func(action *action.BroadcastAction) {
			zap.L().Sugar().Warn("Received broadcast action with no handler!", "action", action.Action, "data", action.Data)
		}
	}

	zap.L().Sugar().Infow("Connected to other nodes", "peers", client.NumMembers(), "ip", nodeInfo.PeerIpAddresses)

	memberList := &Memberlist{
		client:       client,
		delegate:     msgDelegate,
		shutdownChan: make(chan bool),
		onAction:     opts.OnAction,
	}

	memberList.ListenForBroadcastActions()
	return memberList, nil
}

func (m *Memberlist) Broadcast(broadcast *action.BroadcastAction) {
	zap.L().Sugar().Infow("Broadcasting action", "action", broadcast.Action, "data", broadcast.Data)
	m.delegate.Broadcasts.QueueBroadcast(broadcast)
}

func (m *Memberlist) ListenForBroadcastActions() {
	go func() {
		for {
			select {
			case msg := <-m.delegate.MessageChan:
				action, err := action.ParseBroadcastAction(msg)
				if err != nil {
					zap.L().Sugar().Warnw("Failed to parse broadcast action", "err", err, "data", string(msg))
					continue
				}
				go m.onAction(action)
			case <-m.shutdownChan:
				return
			}
		}
	}()
}

func (m *Memberlist) Shutdown() {
	m.client.Shutdown()
	m.shutdownChan <- true
}
