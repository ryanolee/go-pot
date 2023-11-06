package generator

import (
	"embed"
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
	ConfigGeneratorCollection struct {
		generators map[string]*chaff.RootGenerator
	}
)

func NewConfigGeneratorCollection() (*ConfigGeneratorCollection, error) {
	schemaDir, err := schemaFiles.ReadDir(schemaDir)

	if err != nil {
		return nil, err
	}

	logger := zap.L().Sugar()
	generators := make(map[string]*chaff.RootGenerator)
	for _, dirEntry := range schemaDir {
		logger.Debug("Parsing Schema File", "filename", dirEntry.Name())

		generator, err := parseSchemaFile(dirEntry)
		if err != nil {
			logger.Warnw("Failed to parse schema file", "filename", dirEntry.Name(), "error", err)
			continue
		}

		for path, err := range generator.Metadata.Errors {
			logger.Debug("Issue when parsing schema file", "filename", dirEntry.Name(), "path", path, "error", err)
		}

		generators[dirEntry.Name()] = generator
	}

	return &ConfigGeneratorCollection{
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

func (g *ConfigGeneratorCollection) GetRandomGenerator() *chaff.RootGenerator {
	rnd := rand.NewRandUtilFromTime()
	key := funk.Keys(g.generators).([]string)[rnd.RandomInt(0, len(g.generators))]

	return g.generators[key]
}
