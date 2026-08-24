package helpers

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logpath 日志文件路径
// loglevel 日志级别
func initLogger(logpath string, loglevel string) *zap.Logger {

	hook := lumberjack.Logger{
		Filename:   logpath, // 日志文件路径
		MaxSize:    1024,    // megabytes
		MaxBackups: 3,       // 最多保留3个备份
		MaxAge:     7,       //days
		Compress:   true,    // 是否压缩 disabled by default
	}
	w := zapcore.AddSync(&hook)

	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)

	logger := zap.New(core)
	logger.Info("DefaultLogger init success")

	return logger
}

type LoggerWrapper struct {
	ZapLogger *zap.Logger
}

var loggerInstance *LoggerWrapper
var loggerOnce sync.Once

func NewLogger() *LoggerWrapper {
	obj := new(LoggerWrapper)
	return obj
}

func GetLoggerInstance() *LoggerWrapper {
	loggerOnce.Do(func() {
		loggerInstance = NewLogger()
	})
	return loggerInstance
}

func (lg *LoggerWrapper) InitLogger(logpath, loglevel string) {
	if lg.ZapLogger == nil {
		lg.ZapLogger = initLogger(logpath, loglevel)
	}
}
