package gossip

import (
	"strconv"
	"strings"
	"time"

	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"go.uber.org/zap"
)

type (
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
		zap.L().Sugar().Infow("Received ADD_COLD_IP action", "data", action.Data)
		data := strings.Split(action.Data, ",")
		ip := data[0]
		duration, err := strconv.Atoi(data[1])
		if err != nil {
			zap.L().Sugar().Errorw("Failed to parse duration", "data", action.Data)
			return
		}

		h.timeoutWatcher.CommitToColdCache(ip, time.Duration(duration))
	default:
		zap.L().Sugar().Warnw("Received unknown action", "action", action.Action, "data", action.Data)
	}
}
