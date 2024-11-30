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

		// Internals
		minimumRecastInterval int
		maximumRecastInterval int
		timeWastedRatio       float64
	}
)

func NewRecast(lf fx.Lifecycle, shutdowner fx.Shutdowner, config *config.Config, telemetry *Telemetry) (*Recast, error) {
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

		minimumRecastInterval: config.Recast.MinimumRecastIntervalMin,
		maximumRecastInterval: config.Recast.MaximumRecastIntervalMin,
		timeWastedRatio:       config.Recast.TimeWastedRatio,
	}

	// Terminate the recast checker when the application is shutting down
	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			recast.StartChecking()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			recast.shutdownChan <- true
			close(recast.shutdownChan)
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
			// Sleep for a random amount of time between the minimum and maximum recast interval
			recastWaitTime := rand.RandomInt(r.minimumRecastInterval, r.minimumRecastInterval)
			zap.L().Sugar().Info("Recast Waiting", "timeUntilNextCheck", recastWaitTime)
			recastCheckDuration := time.Duration(recastWaitTime) * time.Minute
			select {
			case <-time.After(recastCheckDuration):
				wastedTimeSinceLastCheck := r.telemetry.GetWastedTime() - cumulativeWastedTime
				zap.L().Sugar().Infow("Checking if node should recast", "wastedTimeSinceLastCheck", wastedTimeSinceLastCheck, "timeWastedRatio", timeWastedRatio, "recastCheckDuration", recastCheckDuration)

				if wastedTimeSinceLastCheck < recastCheckDuration.Seconds()*timeWastedRatio {
					zap.L().Sugar().Warnw("Node should recast", "wastedTimeSinceLastCheck", wastedTimeSinceLastCheck, "timeWastedRatio", timeWastedRatio, "recastCheckDuration", recastCheckDuration)
					if err := r.shutdowner.Shutdown(); err != nil {
						// Not sure how to handle this error
						return
					}
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
