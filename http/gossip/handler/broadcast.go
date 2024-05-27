package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"go.uber.org/zap"
)

type (
	IBroadcastActionHandler interface {
		Handle(*action.BroadcastAction)
	}

	BroadcastActionHandler struct {
		timeoutWatcher *metrics.TimeoutWatcher
	}
)

func NewBroadcastActionHandler(timeoutWatcher *metrics.TimeoutWatcher) *BroadcastActionHandler {
	return &BroadcastActionHandler{
		timeoutWatcher: timeoutWatcher,
	}
}

func (h *BroadcastActionHandler) Handle(action *action.BroadcastAction) {
	switch action.Action {
	case "ADD_COLD_IP":
		data := strings.Split(action.Data, ",")
		ip := data[0]
		duration, err := strconv.Atoi(data[1])
		if err != nil {
			zap.L().Sugar().Errorw("Failed to parse duration", "data", action.Data)
			return
		}

		if !h.timeoutWatcher.HasColdCacheTimeout(ip) {
			zap.L().Sugar().Infow("ADD_COLD_IP is new to this node, Rebroadcasting.", "ip", ip, "duration", duration)
			h.timeoutWatcher.CommitToColdCacheWithBroadcast(ip, time.Duration(duration))
		} else {
			h.timeoutWatcher.CommitToColdCache(ip, time.Duration(duration))
		}

	default:
		zap.L().Sugar().Warnw("Received unknown action", "action", action.Action, "data", action.Data)
	}
}
