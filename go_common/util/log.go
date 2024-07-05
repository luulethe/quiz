package util

import (
	"context"
	"os"

	"github.com/luulethe/quiz/go_common/log"
)

func InitLog(ctx context.Context, logPath string, debugLog bool, consoleLog bool, outputFileConfig log.FileConfig) context.Context {
	logLevel := log.InfoLevel
	if debugLog {
		logLevel = log.DebugLevel
	}
	if logPath == "" {
		logPath = "app/logs"
	}

	var err error

	ctx, err = log.Configure(ctx, log.Config{
		Level:                 logLevel,
		EncodeLogsAsJSON:      false,
		ConsoleLoggingEnabled: consoleLog,
		FileLoggingEnabled:    true,
		Directory:             logPath,
		CallerEnabled:         true,
		CallerSkip:            1,
		FileConfig:            outputFileConfig,
		MaxSize:               128, /*MB*/
		MaxBackups:            60})

	ExitOnErr(ctx, err)
	ctx = log.WithFields(ctx, log.Fields{"pid": os.Getpid()})
	log.SetDefaultContext(ctx)
	return ctx
}
