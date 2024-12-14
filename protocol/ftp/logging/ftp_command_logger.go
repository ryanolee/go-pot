package logging

import (
	"net"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"go.uber.org/zap"
)

type (
	CommandLogger interface {
		Log(command string, fields ...zap.Field)
	}

	FtpCommandLogger struct {
		logger                *zap.Logger
		commandsToLog         map[string]bool
		additionalFieldsToLog []string
	}

	contextBoundFtpCommandLogger struct {
		logger  *FtpCommandLogger
		context ftpserver.ClientContext
	}

	contextFieldAccessor func(ftpserver.ClientContext) zap.Field
)

func getPort(addr net.Addr) uint {
	switch v := addr.(type) {
	case *net.UDPAddr:
		return uint(v.Port)
	case *net.TCPAddr:
		return uint(v.Port)
	}

	return 0
}

func getHost(addr net.Addr) string {
	switch v := addr.(type) {
	case *net.UDPAddr:
		return v.IP.String()
	case *net.TCPAddr:
		return v.IP.String()
	}

	return ""
}

var (
	contextFieldAccessors = map[string]contextFieldAccessor{
		"id": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.Uint32("id", ctx.ID())
		},
		"dest_addr": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("dest_addr", ctx.LocalAddr().String())
		},
		"src_addr": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("src_addr", ctx.RemoteAddr().String())
		},
		"dest_port": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.Uint("dest_port", getPort(ctx.LocalAddr()))
		},
		"src_port": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.Uint("src_port", getPort(ctx.RemoteAddr()))
		},
		"dest_host": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("dest_host", getHost(ctx.LocalAddr()))
		},
		"src_host": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("src_host", getHost(ctx.RemoteAddr()))
		},
		"client_version": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("client_version", ctx.GetClientVersion())
		},
		"type": func(ctx ftpserver.ClientContext) zap.Field {
			return zap.String("type", "ftp")
		},
	}
	// Commands that are not included in the "all" command group
	// but are verbose enough to be included in the "all_detailed" group
	overlyVerboseCommands = map[string]bool{
		"read_file":     true,
		"read_file_at":  true,
		"write_file":    true,
		"write_file_at": true,
	}
)

func NewFtpCommandLogger(config *config.Config) (*FtpCommandLogger, error) {
	loggerCfg := zap.NewProductionConfig()

	if config.FtpServer.CommandLog.Path != "" {
		loggerCfg.OutputPaths = []string{config.FtpServer.CommandLog.Path}
	} else if config.Logging.Path != "" {
		loggerCfg.OutputPaths = []string{config.Logging.Path}
	} else {
		loggerCfg.OutputPaths = []string{"stdout"}
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, err
	}

	commandsToLog := make(map[string]bool, len(config.FtpServer.CommandLog.CommandsToLog))
	for _, command := range config.FtpServer.CommandLog.CommandsToLog {
		commandsToLog[command] = true
	}

	return &FtpCommandLogger{
		logger:                logger,
		commandsToLog:         commandsToLog,
		additionalFieldsToLog: config.FtpServer.CommandLog.AdditionalFields,
	}, nil
}

func (l *FtpCommandLogger) ShouldLog(command string) bool {
	// Log commands that are explicitly listed
	if _, ok := l.commandsToLog[command]; ok {
		return true
	}

	// Log all commands if the "all_detailed" command is listed
	if _, ok := l.commandsToLog["all_detailed"]; ok {
		return true
	}

	// Log commands that are not overly verbose if the "all" command group is listed
	_, logAll := l.commandsToLog["all"]
	_, commandIsVerbose := overlyVerboseCommands[command]
	if logAll && !commandIsVerbose {
		return true
	}

	return false
}

func (l *FtpCommandLogger) injectContext(ctx ftpserver.ClientContext, fields []zap.Field) []zap.Field {
	for _, accessor := range l.additionalFieldsToLog {
		if fieldAccessor, ok := contextFieldAccessors[accessor]; ok {
			fields = append(fields, fieldAccessor(ctx))
		}

	}

	return fields
}

func (l *FtpCommandLogger) LogWithContext(ctx ftpserver.ClientContext, command string, fields ...zap.Field) {
	if !l.ShouldLog(command) {
		return
	}

	fields = l.injectContext(ctx, fields)
	l.logger.Info(command, fields...)
}

func (l *FtpCommandLogger) WithContext(context ftpserver.ClientContext) *contextBoundFtpCommandLogger {
	return &contextBoundFtpCommandLogger{
		logger:  l,
		context: context,
	}
}

func (l *contextBoundFtpCommandLogger) Log(command string, fields ...zap.Field) {
	l.logger.LogWithContext(l.context, command, fields...)
}
