package driver

import (
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type FtpClientDriver struct {
	ctx *ftpserver.ClientContext
}

func (f *FtpClientDriver) Create(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__ Create", name)
	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Mkdir(name string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__ Mkdir", name, perm)
	return nil
}

func (f *FtpClientDriver) MkdirAll(path string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__ MkdirAll", path, perm)
	return nil
}

func (f *FtpClientDriver) Open(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__ Open", name)
	return &FtpFile{}, nil
}

func (f *FtpClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__ OpenFile", name, flag, perm)
	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Remove(name string) error {
	zap.L().Sugar().Info("__STUB__ Remove", name)
	return nil
}

func (f *FtpClientDriver) RemoveAll(path string) error {
	zap.L().Sugar().Info("__STUB__ RemoveAll", path)
	return nil
}

func (f *FtpClientDriver) Rename(oldname, newname string) error {
	zap.L().Sugar().Info("__STUB__ Rename", oldname, newname)
	return nil
}

func (f *FtpClientDriver) Stat(name string) (os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__ Stat", name)
	return nil, nil
}

func (f *FtpClientDriver) Name() string {
	zap.L().Sugar().Info("__STUB__ Name")
	return "FtpClientDriver"
}

func (f *FtpClientDriver) Chmod(name string, mode os.FileMode) error {
	zap.L().Sugar().Info("__STUB__ Chmod", name, mode)
	return nil
}

func (f *FtpClientDriver) Chown(name string, uid, gid int) error {
	zap.L().Sugar().Info("__STUB__ Chown", name, uid, gid)
	return nil
}

func (f *FtpClientDriver) Chtimes(name string, atime time.Time, mtime time.Time) error {
	zap.L().Sugar().Info("__STUB__ Chtimes", name, atime, mtime)
	return nil
}





