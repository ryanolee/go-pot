package generator

import (
	"github.com/ryanolee/go-pot/generator/encoder"
	"github.com/ryanolee/go-pot/generator/source"
	"go.uber.org/zap"
)

type (
	TabularGenerator struct {
		encoder encoder.Encoder
	}
)

func NewTabularGenerator(encoder encoder.Encoder) *TabularGenerator {
	return &TabularGenerator{
		encoder: encoder,
	}
}

func (g *TabularGenerator) Generate() []byte {
	data := source.GetTabularFieldValues()

	marshalledData, err := g.encoder.Marshal(data)
	if err != nil {
		zap.L().Sugar().Error(err)
		return nil
	}

	return marshalledData
}

func (g *TabularGenerator) Start() []byte {
	return []byte(g.encoder.Start())
}

func (g *TabularGenerator) GenerateChunk() []byte {
	return g.Generate()
}

func (g *TabularGenerator) ChunkSeparator() []byte {
	return []byte(g.encoder.Delimiter())
}

func (g *TabularGenerator) End() []byte {
	return []byte(g.encoder.End())
}
