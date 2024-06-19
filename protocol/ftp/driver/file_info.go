package driver

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type FtpFileInfo struct {
	path string
}

func NewFtpFileInfo(path string) *FtpFileInfo {
	return &FtpFileInfo{
		path: path,
	}
}

func (f *FtpFileInfo) Name() string {
	zap.L().Sugar().Info("__STUB__ Name")
	return f.path
}

func (f *FtpFileInfo) Size() int64 {
	zap.L().Sugar().Info("__STUB__ Size")
	return 0
}

func (f *FtpFileInfo) Mode() os.FileMode {
	zap.L().Sugar().Info("__STUB__ Mode")
	return 0
}

func (f *FtpFileInfo) ModTime() time.Time {
	zap.L().Sugar().Info("__STUB__ ModTime")
	return time.Now()
}

func (f *FtpFileInfo) IsDir() bool {
	zap.L().Sugar().Info("__STUB__ IsDir")
	fmt.Println(strings.HasSuffix(f.path, "/"))
	return strings.HasSuffix(f.path, "/")
}

func (f *FtpFileInfo) Sys() any {
	zap.L().Sugar().Info("__STUB__ Sys")
	return nil
}
