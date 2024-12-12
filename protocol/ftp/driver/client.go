package driver

import (
	"errors"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/generator/filesystem"
	"github.com/ryanolee/go-pot/protocol/ftp/di"
	"github.com/ryanolee/go-pot/protocol/ftp/logging"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// FTP Client handles top level operations for a single client
// FTP Sever Driver --> [FTP Client Driver] --> FTP File, FTP File Info

type FtpClientDriver struct {
	id        *int64
	ctx       ftpserver.ClientContext
	generator *filesystem.FilesystemGenerator
	repo      *di.FtpRepository
	logger    logging.CommandLogger
}

func NewFtpClientDriver(id *int64, ctx ftpserver.ClientContext, repo *di.FtpRepository) *FtpClientDriver {

	return &FtpClientDriver{
		id:        id,
		ctx:       ctx,
		generator: filesystem.NewFilesystemGenerator(*id),
		repo:      repo,
		logger:    repo.GetLogger().WithContext(ctx),
	}
}

func (f *FtpClientDriver) Create(name string) (afero.File, error) {
	f.logger.Log("create_file", zap.String("path", name))
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return NewFtpFile(name, f.generator, f.ctx, f.repo, f.logger), nil
}

func (f *FtpClientDriver) Mkdir(name string, perm os.FileMode) error {
	f.logger.Log("create_directory", zap.String("path", name), zap.String("perm", perm.String()))

	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) MkdirAll(path string, perm os.FileMode) error {
	f.logger.Log("create_directory_recursive", zap.String("path", path), zap.String("perm", perm.String()))
	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) Open(name string) (afero.File, error) {
	f.logger.Log("open", zap.String("path", name))
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return NewFtpFile(name, f.generator, f.ctx, f.repo, f.logger), nil
}

func (f *FtpClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	f.logger.Log("open_file", zap.String("path", name), zap.Int("flag", flag), zap.String("perm", perm.String()))
	err := f.waitForThrottle()
	if err != nil {
		return nil, err
	}

	return NewFtpFile(name, f.generator, f.ctx, f.repo, f.logger), nil
}

func (f *FtpClientDriver) Remove(name string) error {
	f.logger.Log("remove", zap.String("path", name))

	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) RemoveAll(path string) error {
	f.logger.Log("remove_all", zap.String("path", path))
	err := f.waitForThrottle()
	if err != nil {
		return err
	}

	return nil
}

func (f *FtpClientDriver) Rename(oldname, newname string) error {
	f.logger.Log("rename", zap.String("path", oldname), zap.String("new_path", newname))
	err := f.waitForThrottle()
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Stat(name string) (os.FileInfo, error) {
	err := f.waitForThrottle()
	f.logger.Log("stat", zap.String("path", name))

	if err != nil {
		return nil, err
	}

	fileSize := f.repo.GetConfig().FtpServer.Transfer.FileSize

	return NewFtpFileInfo(name, fileSize), nil
}

func (f *FtpClientDriver) Name() string {
	err := f.waitForThrottle()
	if err != nil {
		return ""
	}

	return "FtpClientDriver"
}

func (f *FtpClientDriver) Chmod(name string, mode os.FileMode) error {
	err := f.waitForThrottle()
	f.logger.Log("chmod", zap.String("path", name), zap.String("mode", mode.String()))
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Chown(name string, uid, gid int) error {
	err := f.waitForThrottle()
	f.logger.Log("chown", zap.String("path", name), zap.Int("uid", uid), zap.Int("gid", gid))
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) Chtimes(name string, atime time.Time, mtime time.Time) error {
	err := f.waitForThrottle()
	f.logger.Log("chtimes", zap.String("path", name), zap.Time("atime", atime), zap.Time("mtime", mtime))
	if err != nil {
		return err
	}
	return nil
}

func (f *FtpClientDriver) waitForThrottle() error {
	waitChannel, err := f.repo.GetThrottle().Throttle(*f.id)

	if err != nil {
		return err
	}

	res := <-waitChannel

	if !res {
		return errors.New("throttle ended unexpectedly")
	}

	return nil
}
