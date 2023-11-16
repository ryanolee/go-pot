package metrics

import (
	"errors"
	"time"

	"github.com/ryanolee/ryan-pot/rand"
	"go.uber.org/zap"
)

const (
	minimumRecastInterval = 30
	maximumRecastInterval = 120

	// If the node has spent less than 5% of its time wasting other nodes time then it should recast
	timeWastedRatio = 0.05
)

// This struct monitors the amount of time wasted by the node and determines if it should try to shutdown the node and
// "recast" to get a new IP address
type (
	RecastInput struct {
		Telemetry *Telemetry
		OnRecast  func()
	}
	Recast struct {
		telemetry    *Telemetry
		shutdownChan chan bool
		onRecast     func()
	}
)

func NewRecast(recastInput *RecastInput) (*Recast, error) {
	if recastInput.Telemetry == nil {
		return nil, errors.New("telemetry is nil")
	}

	if recastInput.OnRecast == nil {
		return nil, errors.New("onRecast is nil")
	}

	return &Recast{
		shutdownChan: make(chan bool),
		telemetry:    recastInput.Telemetry,
		onRecast:     recastInput.OnRecast,
	}, nil
}

func (r *Recast) StartChecking() {
	go func() {
		rand := rand.NewSeededRandFromTime()
		cumulativeWastedTime := 0.0
		for {
			recastCheckDuration := time.Duration(rand.RandomInt(minimumRecastInterval, maximumRecastInterval)) * time.Minute
			select {
			case <-time.After(recastCheckDuration):
				wastedTimeSinceLastCheck := r.telemetry.GetWastedTime() - cumulativeWastedTime
				zap.L().Sugar().Infow("Checking if node should recast", "wastedTimeSinceLastCheck", wastedTimeSinceLastCheck, "timeWastedRatio", timeWastedRatio, "recastCheckDuration", recastCheckDuration)

				if wastedTimeSinceLastCheck < recastCheckDuration.Seconds()*timeWastedRatio {
					zap.L().Sugar().Warnw("Node should recast", "wastedTimeSinceLastCheck", wastedTimeSinceLastCheck, "timeWastedRatio", timeWastedRatio, "recastCheckDuration", recastCheckDuration)
					r.onRecast()
					return
				}

				cumulativeWastedTime = r.telemetry.GetWastedTime()
			case <-r.shutdownChan:
				zap.L().Sugar().Warnw("Shutting down recast checker!")
				return
			}
		}
	}()
}

func (r *Recast) Shutdown() {
	r.shutdownChan <- true
}
