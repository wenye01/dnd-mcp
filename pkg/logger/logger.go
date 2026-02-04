// Package logger 提供日志功能
package logger

import (
	"log"
	"os"
)

// Logger 日志接口
type Logger interface {
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
}

// defaultLogger 默认日志实现
type defaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// Instance 默认日志实例
var Instance Logger = &defaultLogger{
	infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
	errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
}

// Info 信息日志
func (l *defaultLogger) Info(msg string, keyvals ...interface{}) {
	l.infoLogger.Println(append([]interface{}{msg}, keyvals...)...)
}

// Error 错误日志
func (l *defaultLogger) Error(msg string, keyvals ...interface{}) {
	l.errorLogger.Println(append([]interface{}{msg}, keyvals...)...)
}

// Debug 调试日志
func (l *defaultLogger) Debug(msg string, keyvals ...interface{}) {
	l.debugLogger.Println(append([]interface{}{msg}, keyvals...)...)
}

// 提供便捷函数
func Info(msg string, keyvals ...interface{}) {
	Instance.Info(msg, keyvals...)
}

func Error(msg string, keyvals ...interface{}) {
	Instance.Error(msg, keyvals...)
}

func Debug(msg string, keyvals ...interface{}) {
	Instance.Debug(msg, keyvals...)
}

func Fatal(msg string, keyvals ...interface{}) {
	Instance.Error(msg, keyvals...)
	os.Exit(1)
}
