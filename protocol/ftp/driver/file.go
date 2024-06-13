package driver

import (
	"os"

	"go.uber.org/zap"
)

// Close() error
// Read(p []byte) (n int, err error)
// ReadAt(p []byte, off int64) (n int, err error)
// Seek(offset int64, whence int) (int64, error)
// Write(p []byte) (n int, err error)
// WriteAt(p []byte, off int64) (n int, err error)
// Name() string
// Readdir(count int) ([]os.FileInfo, error)
// Readdirnames(n int) ([]string, error)
// Stat() (os.FileInfo, error)
// Sync() error
// Truncate(size int64) error
// WriteString(s string) (ret int, err error)

type FtpFile struct{
}

func (f *FtpFile) Close() error {
	zap.L().Sugar().Info("__STUB__ Close")
	return nil
}

func (f *FtpFile) Read(p []byte) (n int, err error) {
	zap.L().Sugar().Info("__STUB__ Read", p)
	return 0, nil
}

func (f *FtpFile) ReadAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Info("__STUB__ ReadAt", p, off)
	return 0, nil
}

func (f *FtpFile) Seek(offset int64, whence int) (int64, error) {
	zap.L().Sugar().Info("__STUB__ Seek", offset, whence)
	return 0, nil
}

func (f *FtpFile) Write(p []byte) (n int, err error) {
	zap.L().Sugar().Info("__STUB__ Write", p)
	return 0, nil
}

func (f *FtpFile) WriteAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Info("__STUB__ WriteAt", p, off)
	return 0, nil
}

func (f *FtpFile) Name() string {
	zap.L().Sugar().Info("__STUB__ Name")
	return "FtpFile"
}

func (f *FtpFile) Readdir(count int) ([]os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__ Readdir", count)

	return nil, nil
}

func (f *FtpFile) Readdirnames(n int) ([]string, error) {
	zap.L().Sugar().Info("__STUB__ Readdirnames", n)
	return nil, nil
}

func (f *FtpFile) Stat() (os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__ Stat")
	return &FtpFileInfo{}, nil
}

func (f *FtpFile) Sync() error {
	zap.L().Sugar().Info("__STUB__ Sync")
	return nil
}

func (f *FtpFile) Truncate(size int64) error {
	zap.L().Sugar().Info("__STUB__ Truncate", size)
	return nil
}

func (f *FtpFile) WriteString(s string) (ret int, err error) {
	zap.L().Sugar().Info("__STUB__ WriteString", s)
	return 0, nil
}
