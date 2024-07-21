package di

import (
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/generator"
	ftpStall "github.com/ryanolee/ryan-pot/protocol/ftp/stall"
	"github.com/ryanolee/ryan-pot/protocol/ftp/throttle"
	"github.com/ryanolee/ryan-pot/secrets"
)

// Repository used for passing Dependencies required by
// client elements of the FTP protocol
type (
	FtpRepository struct {
		config           *config.Config
		configGenerators *generator.ConfigGeneratorCollection
		secretGenerators *secrets.SecretGeneratorCollection
		throttle         *throttle.FtpThrottle
		stallPool        *stall.StallerPool
		ftpStallFactory  *ftpStall.FtpFileStallerFactory
	}
)

func NewFtpRepository(
	config *config.Config,
	configGenerators *generator.ConfigGeneratorCollection,
	secretGenerators *secrets.SecretGeneratorCollection,
	throttle *throttle.FtpThrottle,
	stallPool *stall.StallerPool,
	ftpStallFactory *ftpStall.FtpFileStallerFactory,
) *FtpRepository {
	return &FtpRepository{
		config:           config,
		configGenerators: configGenerators,
		secretGenerators: secretGenerators,
		throttle:         throttle,
		stallPool:        stallPool,
		ftpStallFactory:  ftpStallFactory,
	}
}

func (r *FtpRepository) GetConfigGenerators() *generator.ConfigGeneratorCollection {
	return r.configGenerators
}

func (r *FtpRepository) GetSecretGenerators() *secrets.SecretGeneratorCollection {
	return r.secretGenerators
}

func (r *FtpRepository) GetThrottle() *throttle.FtpThrottle {
	return r.throttle
}

func (r *FtpRepository) GetStallPool() *stall.StallerPool {
	return r.stallPool
}

func (r *FtpRepository) GetFtpStallFactory() *ftpStall.FtpFileStallerFactory {
	return r.ftpStallFactory
}

func (r *FtpRepository) GetConfig() *config.Config {
	return r.config
}
