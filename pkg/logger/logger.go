// Package logger 提供结构化日志功能
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DEBUG 调试级别
	DEBUG LogLevel = iota
	// INFO 信息级别
	INFO
	// WARN 警告级别
	WARN
	// ERROR 错误级别
	ERROR
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志接口
type Logger interface {
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	With(keyvals ...interface{}) Logger
}

// Config 日志配置
type Config struct {
	Level      LogLevel  // 日志级别
	Format     string    // 输出格式: json, text
	Output     io.Writer // 输出目标
	TimeFormat string    // 时间格式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      INFO,
		Format:     "json",
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
	}
}

// structuredLogger 结构化日志实现
type structuredLogger struct {
	mu         sync.Mutex
	config     *Config
	baseFields map[string]interface{}
}

// New 创建新的日志实例
func New(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}
	return &structuredLogger{
		config:     config,
		baseFields: make(map[string]interface{}),
	}
}

// With 添加基础字段
func (l *structuredLogger) With(keyvals ...interface{}) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.baseFields {
		newFields[k] = v
	}

	// 添加新字段
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			newFields[key] = keyvals[i+1]
		}
	}

	return &structuredLogger{
		config:     l.config,
		baseFields: newFields,
	}
}

// log 内部日志方法
func (l *structuredLogger) log(level LogLevel, msg string, keyvals ...interface{}) {
	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 构建日志字段
	fields := make(map[string]interface{})
	for k, v := range l.baseFields {
		fields[k] = v
	}

	// 添加日志级别和消息
	fields["level"] = level.String()
	fields["message"] = msg
	fields["timestamp"] = time.Now().Format(l.config.TimeFormat)

	// 添加额外字段
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			fields[key] = keyvals[i+1]
		}
	}

	// 输出日志
	if l.config.Format == "json" {
		l.logJSON(fields)
	} else {
		l.logText(fields)
	}
}

// logJSON JSON格式输出
func (l *structuredLogger) logJSON(fields map[string]interface{}) {
	data, err := json.Marshal(fields)
	if err != nil {
		log.Printf("[ERROR] 无法序列化日志: %v", err)
		return
	}
	fmt.Fprintln(l.config.Output, string(data))
}

// logText 文本格式输出
func (l *structuredLogger) logText(fields map[string]interface{}) {
	level, _ := fields["level"].(string)
	msg, _ := fields["message"].(string)
	timestamp, _ := fields["timestamp"].(string)

	output := fmt.Sprintf("[%s] %s %s", timestamp, level, msg)

	// 添加额外字段
	for k, v := range fields {
		if k != "level" && k != "message" && k != "timestamp" {
			output += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	fmt.Fprintln(l.config.Output, output)
}

// Info 信息日志
func (l *structuredLogger) Info(msg string, keyvals ...interface{}) {
	l.log(INFO, msg, keyvals...)
}

// Error 错误日志
func (l *structuredLogger) Error(msg string, keyvals ...interface{}) {
	l.log(ERROR, msg, keyvals...)
}

// Debug 调试日志
func (l *structuredLogger) Debug(msg string, keyvals ...interface{}) {
	l.log(DEBUG, msg, keyvals...)
}

// Warn 警告日志
func (l *structuredLogger) Warn(msg string, keyvals ...interface{}) {
	l.log(WARN, msg, keyvals...)
}

// defaultLogger 默认日志实例（向后兼容）
type defaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// Instance 默认日志实例（全局）
var Instance Logger = New(DefaultConfig())

// Info 信息日志（便捷函数）
func Info(msg string, keyvals ...interface{}) {
	Instance.Info(msg, keyvals...)
}

// Error 错误日志（便捷函数）
func Error(msg string, keyvals ...interface{}) {
	Instance.Error(msg, keyvals...)
}

// Debug 调试日志（便捷函数）
func Debug(msg string, keyvals ...interface{}) {
	Instance.Debug(msg, keyvals...)
}

// Warn 警告日志（便捷函数）
func Warn(msg string, keyvals ...interface{}) {
	Instance.Warn(msg, keyvals...)
}

// Fatal 致命错误日志（便捷函数）
func Fatal(msg string, keyvals ...interface{}) {
	Instance.Error(msg, keyvals...)
	os.Exit(1)
}

// SetConfig 设置全局日志配置
func SetConfig(config *Config) {
	Instance = New(config)
}

// ParseLevel 解析日志级别字符串
func ParseLevel(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return DEBUG
	case "info", "INFO":
		return INFO
	case "warn", "warning", "WARN", "WARNING":
		return WARN
	case "error", "ERROR":
		return ERROR
	default:
		return INFO
	}
}
