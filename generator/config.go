package generator

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ryanolee/go-chaff"
	"github.com/ryanolee/go-chaff/rand"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

var (
	//go:embed data/schema/*
	schemaFiles embed.FS
)

const schemaDir = "data/schema"

type (
	ConfigGenerator struct {
		generators map[string]*chaff.RootGenerator
	}
)

func NewConfigGenerator() (*ConfigGenerator, error) {
	schemaDir, err := schemaFiles.ReadDir(schemaDir)

	if err != nil {
		return nil, err
	}

	logger := zap.L().Sugar()
	generators := make(map[string]*chaff.RootGenerator)
	for _, dirEntry := range schemaDir {
		logger.Infow("Parsing Schema File", "filename", dirEntry.Name())

		generator, err := parseSchemaFile(dirEntry)
		if err != nil {
			logger.Warnw("Failed to parse schema file", "filename", dirEntry.Name(), "error", err)
			continue
		}

		for path, err := range generator.Metadata.Errors {
			logger.Warnw("Issue when parsing schema file", "filename", dirEntry.Name(), "path", path, "error", err)
		}

		generators[dirEntry.Name()] = generator
	}

	return &ConfigGenerator{
		generators: generators,
	}, nil
}

func parseSchemaFile(entry fs.DirEntry) (*chaff.RootGenerator, error) {
	var contents []byte
	var generator chaff.RootGenerator
	var err error

	if entry.IsDir() {
		return nil, fmt.Errorf("failed to parse schema file %s. file is a dir", entry.Name())
	}

	if contents, err = schemaFiles.ReadFile(fmt.Sprintf("%s/%s", schemaDir, entry.Name())); err != nil {
		return nil, err
	}

	if generator, err = chaff.ParseSchemaWithDefaults(contents); err != nil {
		return nil, err
	}

	return &generator, nil
}

func (g *ConfigGenerator) Generate() string {
	rnd := rand.NewRandUtilFromTime()
	key := funk.Keys(g.generators).([]string)[rnd.RandomInt(0, len(g.generators))]

	responseData := g.generators[key].GenerateWithDefaults()
	data, _ := json.Marshal(responseData)

	zap.L().Sugar().Debug("gen", "size", len(data), "file", key)
	return string(data)
}

func (g *ConfigGenerator) GenerateChunk() string {
	return g.Generate()
}
