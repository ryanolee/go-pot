package generator

import (
	"github.com/ryanolee/ryan-pot/http/encoder"
	"github.com/ryanolee/ryan-pot/secrets"
)

type (
	ConfigGenerator struct {
		encoder encoder.Encoder
		generators *ConfigGeneratorCollection
		secrets *secrets.SecretGeneratorCollection
	}
)

func NewConfigGenerator(encoder encoder.Encoder, collection *ConfigGeneratorCollection, secrets *secrets.SecretGeneratorCollection) *ConfigGenerator {
	return &ConfigGenerator{
		encoder: encoder,
		generators: collection,
		secrets: secrets,
	}
}

func (g *ConfigGenerator) Generate() []byte {
	gen := g.generators.GetRandomGenerator().GenerateWithDefaults()
	gen = secrets.InjectSecrets(g.secrets, gen)
	data, err := g.encoder.Marshal(gen)
	if err != nil {
		return nil
	}
		
	return data
}

func (g *ConfigGenerator) Start() []byte {
	return []byte(g.encoder.Start())
}

func (g *ConfigGenerator) GenerateChunk() []byte {
	return  g.Generate()
}

func (g *ConfigGenerator) ChunkSeparator() []byte {
	return []byte(g.encoder.Delimiter())
}

func (g *ConfigGenerator) End() []byte {
	return []byte(g.encoder.End())
}