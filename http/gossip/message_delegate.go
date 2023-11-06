package gossip

import (
	"github.com/hashicorp/memberlist"
)

// Message delegate is used to handle messages from the memberlist
// Based heavily on https://github.com/octu0/example-memberlist/tree/master/05-broadcast
type (
	MessageDelegate struct {
		MessageChan chan []byte
		Broadcasts  *memberlist.TransmitLimitedQueue
	}
	MessageEventDelegate struct {
		Num int
	}
)

func NewMessageEventDelegate() *MessageEventDelegate {
	return &MessageEventDelegate{
		Num: 0,
	}
}
func (d *MessageEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.Num += 1
}
func (d *MessageEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.Num -= 1
}
func (d *MessageEventDelegate) NotifyUpdate(node *memberlist.Node) {
}

func NewMessageDelegate(eventsDelegate *MessageEventDelegate) *MessageDelegate {
	queue := &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return eventsDelegate.Num
		},
		RetransmitMult: 2,
	}
	return &MessageDelegate{
		MessageChan: make(chan []byte),
		Broadcasts:  queue,
	}
}

func (md *MessageDelegate) NotifyMsg(msg []byte) {
	md.MessageChan <- msg
}

func (md *MessageDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return md.Broadcasts.GetBroadcasts(overhead, limit)
}

func (d *MessageDelegate) NodeMeta(limit int) []byte {
	return []byte("")
}

func (d *MessageDelegate) LocalState(join bool) []byte {
	return []byte("")
}

func (d *MessageDelegate) MergeRemoteState(buf []byte, join bool) {}
