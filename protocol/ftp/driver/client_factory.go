package driver

import (
	"math"
	"sync/atomic"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/protocol/ftp/di"
	"github.com/ryanolee/ryan-pot/rand"
)

type FtpClientDriverFactory struct {
	repo             *di.FtpRepository
	clientsConnected *atomic.Int64

	offset int64
}

func NewFtpClientDriverFactory(
	repo *di.FtpRepository,
) *FtpClientDriverFactory {
	random := rand.NewSeededRandFromTime()
	offset := random.Rand.Int63n(math.MaxInt64)

	return &FtpClientDriverFactory{
		repo:             repo,
		clientsConnected: &atomic.Int64{},

		// Random offset for deterministic seeding so that each server restart results in different predictable
		// output. Based on a given client ID
		offset: offset,
	}
}

func (f *FtpClientDriverFactory) FromContext(ctx ftpserver.ClientContext) *FtpClientDriver {
	driverId := f.GetClientIdFromContent(ctx)

	return NewFtpClientDriver(&driverId, ctx, f.repo)
}

func (f *FtpClientDriverFactory) GetClientIdFromContent(ctx ftpserver.ClientContext) int64 {
	return int64(ctx.ID()) + f.offset
}
