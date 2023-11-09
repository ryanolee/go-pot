package stall

import (
	"fmt"
	"sync"
	"time"

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
		MaximumFileDescriptors int
	}
)

func NewHttpStallerPool(opts HttpStallerPoolOptions) *HttpStallerPool {
	return &HttpStallerPool{
		deregisterChan:     make(chan *HttpStaller, opts.MaximumConnections),
		stallers:           NewStallerCollection(),
		maximumConnections: opts.MaximumConnections,
	}
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
			staller.Close()
		}
	}

	time.Sleep(time.Second * 2)
	s.stopChan <- true
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
