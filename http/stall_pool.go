package http

import (
	"fmt"
	"sync"
	"time"
)

type (
	HttpStallerPool struct {
		// Map of connection ID's to map to HttpStaller
		stallers           *StallerCollection
		deregisterChan     chan *HttpStaller
		maximumConnections int
		registrationMutex  sync.Mutex
	}

	HttpStallerPoolOptions struct {
		maximumConnections     int
		maximumFileDescriptors int
	}
)

func NewHttpStallerPool(opts HttpStallerPoolOptions) *HttpStallerPool {
	return &HttpStallerPool{
		deregisterChan:     make(chan *HttpStaller, opts.maximumConnections),
		stallers:           NewStallerCollection(),
		maximumConnections: opts.maximumConnections,
	}
}

func (s *HttpStallerPool) Register(staller *HttpStaller) {
	// Bind pool to staller
	staller.BindPool(s.deregisterChan)

	// Lock Registration
	s.registrationMutex.Lock()
	defer s.registrationMutex.Unlock()

	// Add staller to pool
	s.stallers.Add(staller)

}

func (s *HttpStallerPool) Start() {
	// Deregistration watcher
	go func() {
		for {
			select {
			case staller := <-s.deregisterChan:
				s.stallers.Delete(staller)
			}
		}
	}()

	// Limit watcher
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				s.Prune()
			}
		}
	}()
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
