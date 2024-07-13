package stall

import "github.com/ryanolee/ryan-pot/generator"

type FtpFileStaller struct {
	generator *generator.ConfigGenerator
}

func NewFtpFileStaller() *FtpFileStaller {
	return &FtpFileStaller{}
}
