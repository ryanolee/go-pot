package action

import (
	"encoding/json"

	"github.com/hashicorp/memberlist"
)

type (
	IBroadcastActionDispatcher interface {
		Dispatch(*BroadcastAction)
	}

	BroadcastAction struct {
		Action string `json:"action"`
		Data   string `json:"data"`
	}
)

func (b BroadcastAction) Invalidates(other memberlist.Broadcast) bool {
	return false
}
func (b BroadcastAction) Finished() {
	// nop
}

func (b BroadcastAction) Message() []byte {
	data, err := json.Marshal(b)
	if err != nil {
		return []byte("")
	}
	return data
}

func ParseBroadcastAction(data []byte) (*BroadcastAction, error) {
	action := &BroadcastAction{}
	err := json.Unmarshal(data, action)
	return action, err
}
