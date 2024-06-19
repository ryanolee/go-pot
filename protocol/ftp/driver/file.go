package driver

import (
	"hash/crc64"
	"os"

	"github.com/ryanolee/ryan-pot/generator/filesystem"
	"go.uber.org/zap"
)

type FtpFile struct {
	name       string
	seedOffset int64
	gen        *filesystem.FilesystemGenerator
}

var crc64Table = crc64.MakeTable(crc64.ISO)

func NewFtpFile(name string, gen *filesystem.FilesystemGenerator) *FtpFile {

	return &FtpFile{
		name:       name,
		gen:        gen,
		seedOffset: int64(crc64.Checksum([]byte(name), crc64Table)),
	}
}

func (f *FtpFile) Close() error {
	zap.L().Sugar().Info("__STUB__  FtpFile.Close")
	return nil
}

func (f *FtpFile) Read(p []byte) (n int, err error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Read", p)
	return 0, nil
}

func (f *FtpFile) ReadAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.ReadAt", p, off)
	return 0, nil
}

func (f *FtpFile) Seek(offset int64, whence int) (int64, error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Seek", offset, whence)
	return 0, nil
}

func (f *FtpFile) Write(p []byte) (n int, err error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Write", p)
	return 0, nil
}

func (f *FtpFile) WriteAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.WriteAt", p, off)
	return 0, nil
}

func (f *FtpFile) Name() string {
	zap.L().Sugar().Info("__STUB__  FtpFile.Name")
	return f.name
}

func (f *FtpFile) Readdir(count int) ([]os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Readdir", count)

	f.resetGenerator()
	files := f.gen.GenerateFiles(1000)
	fileInfo := make([]os.FileInfo, 0)
	for _, file := range files {
		fileInfo = append(fileInfo, NewFtpFileInfo(file.Name))
	}

	return fileInfo, nil
}

func (f *FtpFile) Readdirnames(n int) ([]string, error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Readdirnames", n)
	return nil, nil
}

func (f *FtpFile) Stat() (os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.Stat")
	return &FtpFileInfo{}, nil
}

func (f *FtpFile) Sync() error {
	zap.L().Sugar().Info("__STUB__  FtpFile.Sync")
	return nil
}

func (f *FtpFile) Truncate(size int64) error {
	zap.L().Sugar().Info("__STUB__  FtpFile.Truncate", size)
	return nil
}

func (f *FtpFile) WriteString(s string) (ret int, err error) {
	zap.L().Sugar().Info("__STUB__  FtpFile.WriteString", s)
	return 0, nil
}

func (f *FtpFile) resetGenerator() {
	f.gen.ResetWithOffset(f.seedOffset)
}
