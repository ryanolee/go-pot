package driver

import (
	"math"
	"sync/atomic"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/core/metrics"
	"github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/protocol/ftp/throttle"
	"github.com/ryanolee/ryan-pot/rand"
)

type FtpClientDriverFactory struct {
	throttle         *throttle.FtpThrottle
	stallerPool      *stall.StallerPool
	clientsConnected *atomic.Int64
	offset           int64
}

func NewFtpClientDriverFactory(tw *metrics.TimeoutWatcher, sp *stall.StallerPool, throttle *throttle.FtpThrottle) *FtpClientDriverFactory {
	random := rand.NewSeededRandFromTime()
	offset := random.Rand.Int63n(math.MaxInt64)

	return &FtpClientDriverFactory{
		throttle:         throttle,
		stallerPool:      sp,
		clientsConnected: &atomic.Int64{},

		// Random offset for deterministic seeding so that each server restart results in different predictable
		// output. Based on a given client ID
		offset: offset,
	}
}

func (f *FtpClientDriverFactory) FromContext(ctx ftpserver.ClientContext) *FtpClientDriver {
	driverId := f.GetClientIdFromContent(ctx)

	return NewFtpClientDriver(&driverId, ctx, f.throttle)
}

func (f *FtpClientDriverFactory) GetClientIdFromContent(ctx ftpserver.ClientContext) int64 {
	return int64(ctx.ID()) + f.offset
}
