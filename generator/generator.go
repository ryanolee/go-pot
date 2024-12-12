package generator

import (
	"github.com/ryanolee/go-pot/generator/encoder"
	"github.com/ryanolee/go-pot/secrets"
)

type Generator interface {
	Start() []byte
	Generate() []byte
	GenerateChunk() []byte
	ChunkSeparator() []byte
	End() []byte
}

func GetGeneratorForEncoder(encoder encoder.Encoder, configGenerators *ConfigGeneratorCollection, secretsGenerators *secrets.SecretGeneratorCollection) Generator {
	switch encoder.GetSupportedGenerator() {
	case "config":
		return NewConfigGenerator(encoder, configGenerators, secretsGenerators)
	case "tabular":
		return NewTabularGenerator(encoder)
	default:
		return nil
	}
}
