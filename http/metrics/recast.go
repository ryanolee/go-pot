package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/rand"
	"go.uber.org/fx"
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
	Recast struct {
		telemetry    *Telemetry
		shutdownChan chan bool
		shutdowner   fx.Shutdowner
	}
)

func NewRecast(lf fx.Lifecycle, shutdowner fx.Shutdowner, config *config.Config, telemetry *Telemetry ) (*Recast, error) {
	if !config.Recast.Enabled {
		return nil, errors.New("Recast is not enabled")
	}

	if telemetry == nil {
		return nil, errors.New("Telemetry is nil")
	}
	
	recast := &Recast{
		shutdownChan: make(chan bool),
		telemetry:    telemetry,
		shutdowner:   shutdowner,
	}
	

	// Terminate the recast checker when the application is shutting down
	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			recast.Shutdown()
			return nil
		},
	})

	return recast, nil
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
					r.shutdowner.Shutdown()
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
