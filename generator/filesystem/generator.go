package filesystem

import (
	"fmt"

	mRand "math/rand"

	"github.com/ryanolee/ryan-pot/rand"
)

var fileExtensions = []string{
	"txt",
	"sql",
	"csv",
	"toml",
	"json",
	"xml",
	"ini",
	"hcl",
	"tfvars",
}

const (
	minFilesInDir = 5
	maxFilesInDir = 10

	minDirsInDir = 2
	maxDirsInDir = 5

	DirSuffix = "dir"
)

// Semi stable filesystem generator used to generate random filesystem metadata
type (
	FilesystemGenerator struct {
		seed int64
		rand rand.SeededRand
	}

	FilesystemEntry struct {
		Name  string
		IsDir bool
	}
)

// Note that this is not thread safe. Each seeded rand needs to be set per client
func NewFilesystemGenerator(seed int64) *FilesystemGenerator {
	return &FilesystemGenerator{
		seed: seed,
		rand: *rand.NewSeededRand(seed),
	}
}

// Rand funcs
func (fg *FilesystemGenerator) Reset() {
	source := mRand.NewSource(fg.seed)
	fg.rand.SetSource(source)
}

// Resets the generator with a new seed and resets
func (fg *FilesystemGenerator) ResetWithOffset(offset int64) {
	source := mRand.NewSource(fg.seed + offset)
	fg.rand.SetSource(source)
}

// Generate a random directory
func (fg *FilesystemGenerator) GenerateFile() *FilesystemEntry {

	fileName := fg.rand.RandomString(20, rand.AlphabetLower)
	fileExt := fg.rand.StringChoice(&fileExtensions)

	return &FilesystemEntry{
		Name:  fmt.Sprintf("%s.%s", fileName, fileExt),
		IsDir: false,
	}
}

func (fg *FilesystemGenerator) Generate() []*FilesystemEntry {
	filesToGenerate := fg.rand.RandomInt(minFilesInDir, maxFilesInDir)
	dirsToGenerate := fg.rand.RandomInt(minDirsInDir, maxDirsInDir)

	files := make([]*FilesystemEntry, 0)

	for i := 0; i < filesToGenerate; i++ {
		files = append(files, fg.GenerateFile())
	}

	for i := 0; i < dirsToGenerate; i++ {
		files = append(files, fg.GenerateDirectory())
	}

	return files
}

func (fg *FilesystemGenerator) GenerateDirectory() *FilesystemEntry {
	return &FilesystemEntry{
		Name: fg.rand.RandomString(10, rand.AlphabetLower) + DirSuffix + "/",
	}
}

// Filesystem file stringer methods
func (fe *FilesystemEntry) String() string {
	return fmt.Sprintf("File: %s, IsDir: %t", fe.Name, fe.IsDir)
}
