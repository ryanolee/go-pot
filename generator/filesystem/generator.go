package filesystem

import (
	"fmt"
	"time"

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
)

// Semi stable filesystem generator used to generate random filesystem metadata
type (
	FilesystemGenerator struct {
		seed int64
		rand rand.SeededRand
	}

	FilesystemDirectory struct {
		Name string
	}

	FilesystemFile struct {
		Size int64
		Name string
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
func (fg *FilesystemGenerator) GenerateFile(timeToDownload time.Duration) *FilesystemFile {

	fileSize := timeToDownload.Milliseconds()
	fileName := fg.rand.RandomString(20, rand.AlphabetLower)
	fileExt := fg.rand.StringChoice(&fileExtensions)

	return &FilesystemFile{
		Name: fmt.Sprintf("%s.%s", fileName, fileExt),
		Size: fileSize,
	}
}

func (fg *FilesystemGenerator) GenerateFiles(timeToDownload time.Duration) []*FilesystemFile {
	numToGenerate := fg.rand.RandomInt(minFilesInDir, maxFilesInDir)
	files := make([]*FilesystemFile, numToGenerate)

	for i, _ := range files {
		files[i] = fg.GenerateFile(timeToDownload)
	}

	return files
}

// Filesystem file stringer methods

func (ff *FilesystemFile) String() string {
	return fmt.Sprintf("File: %s, Size: %d", ff.Name, ff.Size)
}

// FilesystemDirectory
func (fd *FilesystemDirectory) String() string {
	return fmt.Sprintf("Directory: %s", fd.Name)
}
