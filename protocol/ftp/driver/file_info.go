package driver

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/ryanolee/ryan-pot/generator/filesystem"
	"go.uber.org/zap"
)

var isDirRegexp = regexp.MustCompile(fmt.Sprintf(`(%s\/?|\/|\/\-a)$`, filesystem.DirSuffix))

type FtpFileInfo struct {
	path  string
	isDir bool
}

// FTP Metadata relating to a file
// FTP Sever Driver --> FTP Client Driver --> FTP File, [FTP File Info]

func NewFtpFileInfoFromFsEntry(entry *filesystem.FilesystemEntry) *FtpFileInfo {
	return &FtpFileInfo{
		path:  entry.Name,
		isDir: entry.IsDir,
	}
}

func NewFtpFileInfo(path string) *FtpFileInfo {
	return &FtpFileInfo{
		path: path,
	}
}

func (f *FtpFileInfo) Name() string {
	zap.L().Sugar().Debug("__STUB__ Name")
	return f.path
}

func (f *FtpFileInfo) Size() int64 {
	zap.L().Sugar().Debug("__STUB__ Size")
	return fileSize
}

func (f *FtpFileInfo) Mode() os.FileMode {
	zap.L().Sugar().Debug("__STUB__ Mode")
	return 0
}

func (f *FtpFileInfo) ModTime() time.Time {
	zap.L().Sugar().Debug("__STUB__ ModTime")
	return time.Now()
}

func (f *FtpFileInfo) IsDir() bool {
	zap.L().Sugar().Debug("__STUB__ IsDir")
	return f.isDir ||
		isDirRegexp.MatchString(f.path)
}

func (f *FtpFileInfo) Sys() any {
	zap.L().Sugar().Debug("__STUB__ Sys")
	return nil
}
