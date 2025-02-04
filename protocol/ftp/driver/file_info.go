package driver

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/ryanolee/go-pot/generator/filesystem"
)

var isDirRegexp = regexp.MustCompile(fmt.Sprintf(`(%s\/?|\/|\/\-a)$`, filesystem.DirSuffix))

type FtpFileInfo struct {
	path     string
	fileSize int
}

// FTP Metadata relating to a file
// FTP Sever Driver --> FTP Client Driver --> FTP File, [FTP File Info]

func NewFtpFileInfo(path string, fileSize int) *FtpFileInfo {
	return &FtpFileInfo{
		path:     path,
		fileSize: fileSize,
	}
}

func (f *FtpFileInfo) Name() string {
	return f.path
}

func (f *FtpFileInfo) Size() int64 {
	return int64(f.fileSize)
}

func (f *FtpFileInfo) Mode() os.FileMode {
	return 0
}

func (f *FtpFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *FtpFileInfo) IsDir() bool {
	return isDirRegexp.MatchString(f.path)
}

func (f *FtpFileInfo) Sys() any {
	return nil
}
