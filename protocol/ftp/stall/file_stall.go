package stall

import (
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/generator/encoder"
)

type FtpFileStall struct {
	encoder         *encoder.Encoder
	generator       *generator.Generator
	bytesToGenerate int
}

func NewFtpFileStall(enc *encoder.Encoder, gen *generator.Generator, bytesToGenerate int) *FtpFileStall {
	return &FtpFileStall{
		encoder:         enc,
		generator:       gen,
		bytesToGenerate: bytesToGenerate,
	}
}
