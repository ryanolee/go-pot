package stall

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	HttpStallerPool struct {
		// Map of connection ID's to map to HttpStaller
		stallers           *StallerCollection
		deregisterChan     chan *HttpStaller
		stopChan           chan bool
		maximumConnections int
		registrationMutex  sync.Mutex
	}

	HttpStallerPoolOptions struct {
		MaximumConnections     int
	}
)

func NewHttpStallerPool(lifecycle fx.Lifecycle, config *config.Config) *HttpStallerPool {
	pool := &HttpStallerPool{
		deregisterChan:     make(chan *HttpStaller, config.Staller.MaximumConnections),
		stopChan:           make(chan bool),
		stallers:           NewStallerCollection(),
		maximumConnections: config.Staller.MaximumConnections,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			pool.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			pool.Stop()
			return nil
		},
	})

	return pool
}

func (s *HttpStallerPool) Register(staller *HttpStaller) error {
	if s.stallers.Len() >= s.maximumConnections {
		zap.L().Sugar().Warnw("maximum connections reached, cannot register staller")
		return fmt.Errorf("maximum connections reached, cannot register staller")
	}

	// Bind pool to staller
	staller.BindPool(s.deregisterChan)

	// Lock Registration
	s.registrationMutex.Lock()
	defer s.registrationMutex.Unlock()

	// Add staller to pool
	s.stallers.Add(staller)

	return nil

}

func (s *HttpStallerPool) Start() {
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

func (s *HttpStallerPool) Stop() {
	zap.L().Sugar().Warnw("Stopping staller pool")
	for _, ipMap := range s.stallers.stallers {
		for _, staller := range ipMap {
			zap.L().Sugar().Warnw("Closing staller", "ipAddress", staller.ipAddress, "id", staller.id, "duration", staller.GetElapsedTime())
			staller.Close()
		}
	}

	// Sometimes the deregister channel hangs
	// @todo: Fix this
	go(func() {s.stopChan <- true})()
	
	zap.L().Sugar().Warnw("Stopped staller pool")
}

func (s *HttpStallerPool) Prune() {
	target := int(float64(s.maximumConnections) * 0.9)
	length := s.stallers.Len()

	if length > target {
		diff := length - target
		fmt.Println(diff)
		s.stallers.PruneNByIp(diff)
	}
}
