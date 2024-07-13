package throttle

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type FtpThrottle struct {
	// Map of mutexes used for locking each client's operations
	operationMap map[int64]*sync.Mutex

	// Map of channels used for channels actively hanging calls
	waitChannels map[int64][]chan bool

	// Max number of pending operations
	maxPendingOperations int

	// Time to wait before releasing a pending operation
	waitTime time.Duration

	closeChannel chan bool
}

func NewFtpThrottle(lf fx.Lifecycle, cfg *config.Config) *FtpThrottle {
	if !cfg.FtpServer.Enabled {
		return nil
	}

	throttle := &FtpThrottle{
		operationMap:         make(map[int64]*sync.Mutex),
		waitChannels:         make(map[int64][]chan bool),
		maxPendingOperations: 10,
		waitTime:             time.Second * 1,
	}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			throttle.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			throttle.Close()
			return nil
		},
	})

	return throttle
}

func (t *FtpThrottle) ReleaseOnePendingProcessFromEach() {
	for id := range t.waitChannels {
		t.ReleasePendingProcess(id)
	}
}

func (t *FtpThrottle) ReleaseAll(id int64) {
	// Check if there are any pending operations
	if _, ok := t.waitChannels[id]; !ok {
		return
	}

	lock := t.getLock(id)
	lock.Lock()
	defer lock.Unlock()

	// Release all pending operations
	for _, waitChannel := range t.waitChannels[id] {
		waitChannel <- false
		close(waitChannel)
	}

	delete(t.waitChannels, id)
}

func (t *FtpThrottle) ReleasePendingProcess(id int64) {
	fmt.Println("RELEASE One from", id)

	lock := t.getLock(id)
	lock.Lock()
	defer lock.Unlock()

	// Check if there are any pending operations
	if _, ok := t.waitChannels[id]; !ok {
		return
	}

	if len(t.waitChannels[id]) == 0 {
		return
	}

	// Release the first pending operation
	fmt.Println("RELEASEING One from", id)
	waitChannel := t.waitChannels[id][0]
	waitChannel <- true
	close(waitChannel)

	fmt.Println("Released One from", id)

	// Remove the released operation from the list
	t.waitChannels[id] = t.waitChannels[id][1:]
}

func (t *FtpThrottle) Start() {
	go func() {
		operationTicker := time.Tick(t.waitTime)
		for {
			select {
			case <-t.closeChannel:
				return
			case <-operationTicker:
				t.ReleaseOnePendingProcessFromEach()
			}
		}
	}()
}

// Register a new client and returns a channel to close all pending operations with
func (t *FtpThrottle) Throttle(id int64) (chan bool, error) {
	zap.L().Sugar().Info("Throttling Call ftp client id", id)
	lock := t.getLock(id)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := t.waitChannels[id]; !ok {
		t.waitChannels[id] = make([]chan bool, 0)
	}

	// Handle Too many pending operations
	if len(t.waitChannels[id]) >= t.maxPendingOperations {
		return nil, errors.New("max pending operations reached")
	}

	waitChannel := make(chan bool)
	t.waitChannels[id] = append(t.waitChannels[id], waitChannel)

	return waitChannel, nil
}

func (t *FtpThrottle) Unregister(id int64) {
	t.ReleaseAll(id)
	delete(t.operationMap, id)
}

func (t *FtpThrottle) Close() {
	t.closeChannel <- true
	close(t.closeChannel)
	for id := range t.waitChannels {
		t.ReleaseAll(id)
	}
}

func (t *FtpThrottle) getLock(id int64) *sync.Mutex {
	if _, ok := t.operationMap[id]; !ok {
		t.operationMap[id] = &sync.Mutex{}
	}

	return t.operationMap[id]
}
