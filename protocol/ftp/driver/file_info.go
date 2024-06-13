package driver

import (
	"os"
	"time"

	"go.uber.org/zap"
)

type FtpFileInfo struct {
}

func (f *FtpFileInfo) Name() string {
	zap.L().Sugar().Info("__STUB__ Name")
	return "FtpFileInfo"
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
	return false
}

func (f *FtpFileInfo) Sys() any {
	zap.L().Sugar().Info("__STUB__ Sys")
	return nil
}

