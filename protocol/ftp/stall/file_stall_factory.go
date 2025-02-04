package stall

import (
	"fmt"
	"hash/crc64"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/core/stall"
	"github.com/ryanolee/go-pot/generator"
	"github.com/ryanolee/go-pot/generator/encoder"
	"github.com/ryanolee/go-pot/secrets"
	"go.uber.org/zap"
)

var crc64Table = crc64.MakeTable(crc64.ISO)

type (
	FtpFileStallerFactory struct {
		config           *config.Config
		stallerPool      *stall.StallerPool
		configGenerators *generator.ConfigGeneratorCollection
		secretGenerators *secrets.SecretGeneratorCollection
		logger           *zap.Logger
	}
)

func NewFtpFileStallerFactory(
	config *config.Config,
	stallerPool *stall.StallerPool,
	configGenerators *generator.ConfigGeneratorCollection,
	secretGenerators *secrets.SecretGeneratorCollection,
	logger *zap.Logger,
) *FtpFileStallerFactory {
	return &FtpFileStallerFactory{
		config:           config,
		stallerPool:      stallerPool,
		configGenerators: configGenerators,
		secretGenerators: secretGenerators,
		logger:           logger,
	}
}

// Creates a single file stalling handle
func (f *FtpFileStallerFactory) FromName(ctx ftpserver.ClientContext, name string, size int) *FtpFileStaller {
	encoderInstance := encoder.GetEncoderForPath(name)
	generatorInstance := generator.GetGeneratorForEncoder(encoderInstance, f.configGenerators, f.secretGenerators)
	stallerId := crc64.Checksum([]byte(name), crc64Table)

	staller := NewFtpFileStall(&NewFtpFileStallerArgs{
		Config:      f.config,
		Id:          stallerId,
		GroupId:     fmt.Sprintf("ftp-%d", ctx.ID()),
		Encoder:     encoderInstance,
		Generator:   generatorInstance,
		BytesToSend: size,
	})

	if err := f.stallerPool.Register(staller); err != nil {
		f.logger.Warn("Failed to register staller", zap.Error(err))
		staller.Close()
		return nil
	}

	return staller
}
