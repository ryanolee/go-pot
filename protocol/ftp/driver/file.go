package driver

import (
	"hash/crc64"
	"io"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/generator/filesystem"
	"github.com/ryanolee/go-pot/protocol/ftp/di"
	"github.com/ryanolee/go-pot/protocol/ftp/logging"
	ftpStall "github.com/ryanolee/go-pot/protocol/ftp/stall"

	"go.uber.org/zap"
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

	// Logger
	logger logging.CommandLogger

	// Config
	transferChunkSize int
	transferDelay     time.Duration
	fileSize          int
}

var crc64Table = crc64.MakeTable(crc64.ISO)

func NewFtpFile(name string, gen *filesystem.FilesystemGenerator, ctx ftpserver.ClientContext, repo *di.FtpRepository, logger logging.CommandLogger) *FtpFile {
	fileSize := repo.GetConfig().FtpServer.Transfer.FileSize
	return &FtpFile{
		name:              name,
		gen:               gen,
		seedOffset:        int64(crc64.Checksum([]byte(name), crc64Table)),
		ctx:               ctx,
		stall:             repo.GetFtpStallFactory().FromName(ctx, name, fileSize),
		transferChunkSize: repo.GetConfig().FtpServer.Transfer.ChunkSize,
		transferDelay:     time.Duration(repo.GetConfig().FtpServer.Transfer.ChunkSendRate) * time.Millisecond,
		fileSize:          fileSize,
		logger:            logger,
	}
}

func (f *FtpFile) Close() error {
	f.logger.Log("close_file", zap.String("path", f.name))
	f.stall.Halt()
	return nil
}

func (f *FtpFile) Read(p []byte) (n int, err error) {
	f.logger.Log("read_file", zap.String("path", f.name), zap.Int("data_requested", len(p)))
	time.Sleep(f.transferDelay)
	return f.stall.Read(p)
}

func (f *FtpFile) ReadAt(p []byte, off int64) (n int, err error) {
	f.logger.Log("read_file_at", zap.String("path", f.name), zap.Int64("offset", off), zap.Int("data_requested", len(p)))
	return 0, io.EOF
}

func (f *FtpFile) Seek(offset int64, whence int) (int64, error) {
	f.logger.Log("seek_file", zap.String("path", f.name), zap.Int64("offset", offset), zap.Int("whence", whence))
	return 0, nil
}

func (f *FtpFile) Write(p []byte) (n int, err error) {
	f.logger.Log("write_file", zap.String("path", f.name), zap.Int("data_written", len(p)))
	return 0, nil
}

func (f *FtpFile) WriteAt(p []byte, off int64) (n int, err error) {
	f.logger.Log("write_file_at", zap.String("path", f.name), zap.Int64("offset", off), zap.Int("data_written", len(p)))
	return 0, nil
}

func (f *FtpFile) Name() string {
	return f.name
}

func (f *FtpFile) Readdir(count int) ([]os.FileInfo, error) {
	f.logger.Log("read_dir", zap.String("path", f.name))

	f.resetGenerator()
	files := f.gen.Generate()
	fileInfo := make([]os.FileInfo, 0)
	for _, file := range files {
		fileInfo = append(fileInfo, NewFtpFileInfo(file.Name, f.fileSize))
	}

	return fileInfo, nil
}

func (f *FtpFile) Readdirnames(n int) ([]string, error) {
	f.logger.Log("read_dir_names", zap.String("path", f.name))
	return nil, nil
}

func (f *FtpFile) Stat() (os.FileInfo, error) {
	f.logger.Log("stat", zap.String("path", f.name))
	return &FtpFileInfo{}, nil
}

func (f *FtpFile) Sync() error {
	f.logger.Log("sync", zap.String("path", f.name))
	return nil
}

func (f *FtpFile) Truncate(size int64) error {
	f.logger.Log("truncate", zap.String("path", f.name), zap.Int64("size", size))
	return nil
}

func (f *FtpFile) WriteString(s string) (ret int, err error) {
	f.logger.Log("write_string", zap.String("path", f.name), zap.Int("data_written", len(s)))
	return 0, nil
}

func (f *FtpFile) resetGenerator() {
	f.gen.ResetWithOffset(f.seedOffset)
}
