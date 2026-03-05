package logger

import (
	"bill-management/pkg/config"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2" // 日志切割库
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	once   sync.Once
)

func InitLogger() {
	once.Do(func() {
		// 1. 获取配置
		cfg := config.GetConfig().Log

		//2. 解析日志级别
		level, err := zapcore.ParseLevel(cfg.Level)
		if err != nil {
			level = zapcore.DebugLevel
		}

		// 3. 设置日志编码器
		encodingConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		// 4. 选择编码器
		var encoder zapcore.Encoder
		switch cfg.Format {
		case "json":
			encoder = zapcore.NewJSONEncoder(encodingConfig)
		default:
			encoder = zapcore.NewConsoleEncoder(encodingConfig)
		}
		// 5. 设置日志输出
		// 5.1 日志切割
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackup,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		// 5.2 输出目标
		var writeSyncer zapcore.WriteSyncer
		if cfg.Stdout {
			writeSyncer = zap.CombineWriteSyncers(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
		} else {
			writeSyncer = zapcore.AddSync(lumberJackLogger)
		}
		// 6. 创建核心
		core := zapcore.NewCore(encoder, writeSyncer, level)
		// 7. 创建Logger
		logger = zap.New(
			core,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
			zap.Development(),
		)
		// 8. 创建SugaredLogger (提供更友好的API)
		sugar = logger.Sugar()
		defer logger.Sync() // 确保日志被刷新

		logger.Info("日志模块初始化完成")

	})

}

// ========== 全局调用方法（封装 zap 原生方法） ==========

// Logger 获取 zap.Logger 实例（高性能、结构化）
func Logger() *zap.Logger {
	if logger == nil {
		panic("日志模块未初始化，请先调用 logger.Init()")
	}
	return logger
}

// Sugar 获取 zap.SugaredLogger 实例（易用、格式化）
func Sugar() *zap.SugaredLogger {
	if sugar == nil {
		panic("日志模块未初始化，请先调用 logger.Init()")
	}
	return sugar
}

// ========== 快捷日志方法（简化调用） ==========

// Debug 调试日志（结构化）
func Debug(msg string, fields ...zap.Field) {
	Logger().Debug(msg, fields...)
}

// Info 普通日志（结构化）
func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, fields...)
}

// Warn 警告日志（结构化）
func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, fields...)
}

// Error 错误日志（结构化）
func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, fields...)
}

// Fatal 致命日志（输出后退出程序）
func Fatal(msg string, fields ...zap.Field) {
	Logger().Fatal(msg, fields...)
}

// ========== Sugared 快捷方法（格式化字符串） ==========

// Debugf 调试日志（格式化）
func Debugf(format string, args ...interface{}) {
	Sugar().Debugf(format, args...)
}

// Infof 普通日志（格式化）
func Infof(format string, args ...interface{}) {
	Sugar().Infof(format, args...)
}

// Warnf 警告日志（格式化）
func Warnf(format string, args ...interface{}) {
	Sugar().Warnf(format, args...)
}

// Errorf 错误日志（格式化）
func Errorf(format string, args ...interface{}) {
	Sugar().Errorf(format, args...)
}

// Fatalf 致命日志（格式化，输出后退出）
func Fatalf(format string, args ...interface{}) {
	Sugar().Fatalf(format, args...)
}
