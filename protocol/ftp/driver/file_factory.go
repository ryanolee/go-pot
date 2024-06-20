package driver

type FtpFileFactory struct {
}

func NewFtpFileFactory() *FtpFileFactory {
	return &FtpFileFactory{}
}

func (f *FtpFileFactory) FromClientDriver(driver *FtpClientDriver, name string) *FtpFile {
	return NewFtpFile(name, driver.generator)
}
