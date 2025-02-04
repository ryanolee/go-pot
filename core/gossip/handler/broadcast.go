package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/ryanolee/go-pot/core/gossip/action"
	"github.com/ryanolee/go-pot/core/metrics"
	"go.uber.org/zap"
)

type (
	IBroadcastActionHandler interface {
		Handle(*action.BroadcastAction)
	}

	BroadcastActionHandler struct {
		timeoutWatcher *metrics.TimeoutWatcher
		logger         *zap.Logger
	}
)

func NewBroadcastActionHandler(timeoutWatcher *metrics.TimeoutWatcher, logger *zap.Logger) *BroadcastActionHandler {
	return &BroadcastActionHandler{
		timeoutWatcher: timeoutWatcher,
		logger:         logger,
	}
}

func (h *BroadcastActionHandler) Handle(action *action.BroadcastAction) {
	switch action.Action {
	case "ADD_COLD_IP":
		data := strings.Split(action.Data, ",")
		ip := data[0]
		duration, err := strconv.Atoi(data[1])
		if err != nil {
			h.logger.Sugar().Errorw("Failed to parse duration", "data", action.Data)
			return
		}

		if !h.timeoutWatcher.HasColdCacheTimeout(ip) {
			h.logger.Sugar().Infow("ADD_COLD_IP is new to this node, Rebroadcasting.", "ip", ip, "duration", duration)
			h.timeoutWatcher.CommitToColdCacheWithBroadcast(ip, time.Duration(duration))
		} else {
			h.timeoutWatcher.CommitToColdCache(ip, time.Duration(duration))
		}

	default:
		h.logger.Sugar().Warnw("Received unknown action", "action", action.Action, "data", action.Data)
	}
}
