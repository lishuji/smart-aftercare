package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init 初始化日志
func Init(logDir string) {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)

	// 确保日志目录存在
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			log.Warnf("创建日志目录失败: %v，使用标准输出", err)
			return
		}

		logFile := filepath.Join(logDir, "app.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Warnf("打开日志文件失败: %v，使用标准输出", err)
			return
		}

		// 同时输出到文件和标准输出
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	}
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	return log
}

// Info 日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 格式化日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Debug 调试日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}
