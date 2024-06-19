package driver

import (
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/generator/filesystem"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type FtpClientDriver struct {
	id        *int64
	ctx       ftpserver.ClientContext
	generator *filesystem.FilesystemGenerator
}

func NewFtpClientDriver(id *int64, ctx ftpserver.ClientContext) *FtpClientDriver {
	return &FtpClientDriver{
		id:        id,
		ctx:       ctx,
		generator: filesystem.NewFilesystemGenerator(*id),
	}
}

func (f *FtpClientDriver) Create(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Create", name)
	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Mkdir(name string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Mkdir", name, perm)
	return nil
}

func (f *FtpClientDriver) MkdirAll(path string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.MkdirAll", path, perm)
	return nil
}

func (f *FtpClientDriver) Open(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Open", name)
	return NewFtpFile(name, f.generator), nil
}

func (f *FtpClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.OpenFile", name, flag, perm)
	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Remove(name string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Remove", name)
	return nil
}

func (f *FtpClientDriver) RemoveAll(path string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.RemoveAll", path)
	return nil
}

func (f *FtpClientDriver) Rename(oldname, newname string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Rename", oldname, newname)
	return nil
}

func (f *FtpClientDriver) Stat(name string) (os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Stat", name)
	return NewFtpFileInfo(name), nil
}

func (f *FtpClientDriver) Name() string {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Name")
	return "FtpClientDriver"
}

func (f *FtpClientDriver) Chmod(name string, mode os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chmod", name, mode)
	return nil
}

func (f *FtpClientDriver) Chown(name string, uid, gid int) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chown", name, uid, gid)
	return nil
}

func (f *FtpClientDriver) Chtimes(name string, atime time.Time, mtime time.Time) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chtimes", name, atime, mtime)
	return nil
}
