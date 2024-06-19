package driver

import (
	"math"
	"sync/atomic"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/core/metrics"
	"github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/rand"
)

type FtpClientDriverFactory struct {
	timeoutWatcher   *metrics.TimeoutWatcher
	stallerPool      *stall.StallerPool
	clientsConnected *atomic.Int64
	offset           int64
}

func NewFtpClientDriverFactory(tw *metrics.TimeoutWatcher, sp *stall.StallerPool) *FtpClientDriverFactory {
	random := rand.NewSeededRandFromTime()
	offset := random.Rand.Int63n(math.MaxInt64)

	return &FtpClientDriverFactory{
		timeoutWatcher:   tw,
		stallerPool:      sp,
		clientsConnected: &atomic.Int64{},

		// Random offset for deterministic seeding so that each server restart results in different predictable
		// output. Based on a given client ID
		offset: offset,
	}
}

func (f *FtpClientDriverFactory) FromContext(ctx ftpserver.ClientContext) *FtpClientDriver {
	// Pull new ID and offset it
	driverId := f.clientsConnected.Add(1)
	driverId += f.offset

	return NewFtpClientDriver(&driverId, ctx)
}
