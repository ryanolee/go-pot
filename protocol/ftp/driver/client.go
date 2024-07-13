package driver

import (
	"errors"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/generator/filesystem"
	"github.com/ryanolee/ryan-pot/protocol/ftp/throttle"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// FTP Client handles top level operations for a single client
// FTP Sever Driver --> [FTP Client Driver] --> FTP File, FTP File Info

type FtpClientDriver struct {
	id        *int64
	ctx       *ftpserver.ClientContext
	generator *filesystem.FilesystemGenerator
	throttle  *throttle.FtpThrottle
}

func NewFtpClientDriver(id *int64, ctx ftpserver.ClientContext, throttle *throttle.FtpThrottle) *FtpClientDriver {

	return &FtpClientDriver{
		id:        id,
		ctx:       &ctx,
		generator: filesystem.NewFilesystemGenerator(*id),
		throttle:  throttle,
	}
}

func (f *FtpClientDriver) Create(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Create", name)
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Mkdir(name string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Mkdir", name, perm)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) MkdirAll(path string, perm os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.MkdirAll", path, perm)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) Open(name string) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Open", name)
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return NewFtpFile(name, f.generator, f.ctx), nil
}

func (f *FtpClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.OpenFile", name, flag, perm)
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return &FtpFile{}, nil
}

func (f *FtpClientDriver) Remove(name string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Remove", name)

	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) RemoveAll(path string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.RemoveAll", path)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) Rename(oldname, newname string) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Rename", oldname, newname)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Stat(name string) (os.FileInfo, error) {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Stat", name)

	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return NewFtpFileInfo(name), nil
}

func (f *FtpClientDriver) Name() string {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Name")

	err := f.waitForThrottle()
	if err != nil {
		return ""
	}

	return "FtpClientDriver"
}

func (f *FtpClientDriver) Chmod(name string, mode os.FileMode) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chmod", name, mode)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Chown(name string, uid, gid int) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chown", name, uid, gid)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Chtimes(name string, atime time.Time, mtime time.Time) error {
	zap.L().Sugar().Info("__STUB__  FtpClientDriver.Chtimes", name, atime, mtime)
	err := f.waitForThrottle()
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) waitForThrottle() error {
	waitChannel, err := f.throttle.Throttle(*f.id)

	if err != nil {
		return err
	}

	res := <-waitChannel

	if !res {
		return errors.New("throttle ended unexpectedly")
	}

	return nil
}
