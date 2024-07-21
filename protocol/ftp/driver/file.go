package driver

import (
	"hash/crc64"
	"io"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/generator/filesystem"
	"github.com/ryanolee/ryan-pot/protocol/ftp/di"
	ftpStall "github.com/ryanolee/ryan-pot/protocol/ftp/stall"

	"go.uber.org/zap"
)

const (
	transferChunkSize = 16
	transferDelay     = time.Millisecond * 100
	fileSize          = 1024 * 1024 // 1MB
)

// FTP File handles I/O operations for a single file
// FTP Sever Driver --> FTP Client Driver --> [FTP File], FTP File Info
type FtpFile struct {
	name       string
	seedOffset int64

	// Generator details
	stall *ftpStall.FtpFileStaller

	// Service references
	gen *filesystem.FilesystemGenerator
	ctx ftpserver.ClientContext
}

var crc64Table = crc64.MakeTable(crc64.ISO)

func NewFtpFile(name string, gen *filesystem.FilesystemGenerator, ctx ftpserver.ClientContext, repo *di.FtpRepository) *FtpFile {
	return &FtpFile{
		name:       name,
		gen:        gen,
		seedOffset: int64(crc64.Checksum([]byte(name), crc64Table)),
		ctx:        ctx,
		stall:      repo.GetFtpStallFactory().FromName(ctx, name, fileSize),
	}
}

func (f *FtpFile) Close() error {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Close")
	f.stall.Halt()
	return nil
}

func (f *FtpFile) Read(p []byte) (n int, err error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Read", fileSize)
	time.Sleep(transferDelay)
	return f.stall.Read(p)
}

func (f *FtpFile) ReadAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.ReadAt", fileSize, off)
	return 0, io.EOF
}

func (f *FtpFile) Seek(offset int64, whence int) (int64, error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Seek", offset, whence)
	return 0, nil
}

func (f *FtpFile) Write(p []byte) (n int, err error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Write", p)
	return 0, nil
}

func (f *FtpFile) WriteAt(p []byte, off int64) (n int, err error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.WriteAt", p, off)
	return 0, nil
}

func (f *FtpFile) Name() string {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Name")
	return f.name
}

func (f *FtpFile) Readdir(count int) ([]os.FileInfo, error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Readdir", count)

	f.resetGenerator()
	files := f.gen.Generate()
	fileInfo := make([]os.FileInfo, 0)
	for _, file := range files {
		fileInfo = append(fileInfo, NewFtpFileInfo(file.Name))
	}

	return fileInfo, nil
}

func (f *FtpFile) Readdirnames(n int) ([]string, error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Readdirnames", n)
	return nil, nil
}

func (f *FtpFile) Stat() (os.FileInfo, error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Stat")
	return &FtpFileInfo{}, nil
}

func (f *FtpFile) Sync() error {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Sync")
	return nil
}

func (f *FtpFile) Truncate(size int64) error {
	zap.L().Sugar().Debug("__STUB__  FtpFile.Truncate", size)
	return nil
}

func (f *FtpFile) WriteString(s string) (ret int, err error) {
	zap.L().Sugar().Debug("__STUB__  FtpFile.WriteString", s)
	return 0, nil
}

func (f *FtpFile) resetGenerator() {
	f.gen.ResetWithOffset(f.seedOffset)
}
