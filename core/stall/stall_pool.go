package stall

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ryanolee/go-pot/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	StallerPool struct {
		stallers           *StallerCollection
		logger             *zap.Logger
		deregisterChan     chan Staller
		stopChan           chan bool
		maximumConnections int
		registrationMutex  sync.Mutex
	}

	StallerPoolOptions struct {
		MaximumConnections int
	}
)

func NewStallerPool(lifecycle fx.Lifecycle, config *config.Config, logger *zap.Logger) *StallerPool {
	pool := &StallerPool{
		logger:             logger,
		deregisterChan:     make(chan Staller, config.Staller.MaximumConnections),
		stopChan:           make(chan bool),
		stallers:           NewStallerCollection(config.Staller.GroupLimit),
		maximumConnections: config.Staller.MaximumConnections,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			pool.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			// Top priority is to stop the pool while connections are still active
			// so is called at top level
			return nil
		},
	})

	return pool
}

func (s *StallerPool) Register(staller Staller) error {
	if s.stallers.Len() >= s.maximumConnections {
		s.logger.Sugar().Warnw("maximum connections reached, cannot register staller")
		return fmt.Errorf("maximum connections reached, cannot register staller")
	}

	// Bind pool to staller
	staller.BindToPool(s.deregisterChan)

	// Lock Registration
	s.registrationMutex.Lock()
	defer s.registrationMutex.Unlock()

	// If we fail to add a staller close it and close it and abort the registation the registration
	if err := s.stallers.Add(staller); err != nil {
		staller.Close()
		s.logger.Error("failed to add staller", zap.Error(err))
		return err
	}

	return nil

}

func (s *StallerPool) Start() {
	// Deregistration watcher
	go func() {
		for {
			select {
			case staller := <-s.deregisterChan:
				s.stallers.Delete(staller)
			case <-s.stopChan:
				return
			}
		}
	}()

	// Limit watcher
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				s.Prune()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *StallerPool) Stop() {
	s.logger.Sugar().Warnw("Stopping staller pool")
	for _, ipMap := range s.stallers.stallers {
		for _, staller := range ipMap {
			s.logger.Sugar().Warnw("Closing staller", "group", staller.GetGroupIdentifier(), "id", staller.GetIdentifier())
			staller.Close()
		}
	}

	go (func() { s.stopChan <- true })()

	s.logger.Sugar().Warnw("Stopped staller pool")
}

func (s *StallerPool) StopByIdentifier(id string) {
	s.stallers.PruneByIdentifierGroup(id)
}

func (s *StallerPool) Prune() {
	target := int(float64(s.maximumConnections) * 0.9)
	length := s.stallers.Len()

	if length > target {
		diff := length - target
		s.stallers.PruneNByIdentifier(diff)
	}
}
