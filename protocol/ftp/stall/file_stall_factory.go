package stall

import (
	"fmt"
	"hash/crc64"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/generator/encoder"
	"github.com/ryanolee/ryan-pot/secrets"
)

var crc64Table = crc64.MakeTable(crc64.ISO)

type (
	FtpFileStallerFactory struct {
		config           *config.Config
		stallerPool      *stall.StallerPool
		configGenerators *generator.ConfigGeneratorCollection
		secretGenerators *secrets.SecretGeneratorCollection
	}
)

func NewFtpFileStallerFactory(
	config *config.Config,
	stallerPool *stall.StallerPool,
	configGenerators *generator.ConfigGeneratorCollection,
	secretGenerators *secrets.SecretGeneratorCollection,
) *FtpFileStallerFactory {
	return &FtpFileStallerFactory{
		config:           config,
		stallerPool:      stallerPool,
		configGenerators: configGenerators,
		secretGenerators: secretGenerators,
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

	f.stallerPool.Register(staller)
	return staller
}
