package logger

import (
	"mychat/internal/misc/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFilePath = "app.log"
	maxSizeMB   = 10 // 文件最大尺寸 (MB)
	maxBackups  = 50 // 保留旧文件的最大个数
	maxAgeDays  = 60 // 保留旧文件的最大天数
)

func Init() *zap.Logger {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   false, // 是否启用 gzip 压缩
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	const customTimeLayout = "2006-01-02 15:04:05.000"
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(customTimeLayout)
	if config.Config.DevMode {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	var core zapcore.Core
	if config.Config.DevMode {
		multiWriteSyncer := zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(lumberJackLogger),
			zapcore.AddSync(os.Stdout),
		)

		core = zapcore.NewCore(
			encoder,
			multiWriteSyncer,
			zap.DebugLevel,
		)
	} else {
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(lumberJackLogger),
			zap.InfoLevel,
		)
	}

	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)

	return logger
}
