package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStructuredLogger_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  DEBUG,
		Format: "json",
		Output: buf,
	}
	logger := New(config)

	logger.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// 验证是有效的JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// 验证必需字段
	if logEntry["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", logEntry["level"])
	}
	if logEntry["message"] != "test message" {
		t.Errorf("expected message 'test message', got %v", logEntry["message"])
	}
	if logEntry["key1"] != "value1" {
		t.Errorf("expected key1 value1, got %v", logEntry["key1"])
	}
	if logEntry["key2"] != float64(42) {
		t.Errorf("expected key2 42, got %v", logEntry["key2"])
	}
	if _, ok := logEntry["timestamp"]; !ok {
		t.Error("expected timestamp field")
	}
}

func TestStructuredLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  DEBUG,
		Format: "text",
		Output: buf,
	}
	logger := New(config)

	logger.Info("test message", "key1", "value1")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", output)
	}
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("expected 'key1=value1' in output, got: %s", output)
	}
}

func TestStructuredLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     LogLevel
		msgLevel     LogLevel
		shouldOutput bool
	}{
		{"DEBUG level logs DEBUG", DEBUG, DEBUG, true},
		{"DEBUG level logs INFO", DEBUG, INFO, true},
		{"DEBUG level logs WARN", DEBUG, WARN, true},
		{"DEBUG level logs ERROR", DEBUG, ERROR, true},
		{"INFO level skips DEBUG", INFO, DEBUG, false},
		{"INFO level logs INFO", INFO, INFO, true},
		{"INFO level logs ERROR", INFO, ERROR, true},
		{"WARN level skips INFO", WARN, INFO, false},
		{"WARN level logs WARN", WARN, WARN, true},
		{"ERROR level skips WARN", ERROR, WARN, false},
		{"ERROR level logs ERROR", ERROR, ERROR, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			config := &Config{
				Level:  tt.logLevel,
				Format: "text",
				Output: buf,
			}
			logger := New(config)

			switch tt.msgLevel {
			case DEBUG:
				logger.Debug("debug msg")
			case INFO:
				logger.Info("info msg")
			case WARN:
				logger.Warn("warn msg")
			case ERROR:
				logger.Error("error msg")
			}

			output := buf.String() != ""
			if output != tt.shouldOutput {
				t.Errorf("expected output=%v, got %v", tt.shouldOutput, output)
			}
		})
	}
}

func TestStructuredLogger_With(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  DEBUG,
		Format: "json",
		Output: buf,
	}
	logger := New(config)

	logger = logger.With("service", "test-service", "version", "1.0.0")
	logger.Info("test message")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if logEntry["service"] != "test-service" {
		t.Errorf("expected service test-service, got %v", logEntry["service"])
	}
	if logEntry["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", logEntry["version"])
	}
}

func TestStructuredLogger_WithChaining(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  DEBUG,
		Format: "json",
		Output: buf,
	}
	logger := New(config)

	logger1 := logger.With("base", "field")
	logger2 := logger1.With("extra", "value")

	logger2.Info("test message")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if logEntry["base"] != "field" {
		t.Errorf("expected base field, got %v", logEntry["base"])
	}
	if logEntry["extra"] != "value" {
		t.Errorf("expected extra value, got %v", logEntry["extra"])
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"DEBUG", DEBUG},
		{"info", INFO},
		{"INFO", INFO},
		{"warn", WARN},
		{"WARN", WARN},
		{"warning", WARN},
		{"WARNING", WARN},
		{"error", ERROR},
		{"ERROR", ERROR},
		{"invalid", INFO}, // 默认INFO
		{"", INFO},        // 默认INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel(%d).String() = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != INFO {
		t.Errorf("expected default level INFO, got %v", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("expected default format json, got %v", config.Format)
	}
	if config.TimeFormat == "" {
		t.Error("expected non-empty time format")
	}
}

func TestGlobalLogger(t *testing.T) {
	// 保存原始实例
	original := Instance
	defer func() { Instance = original }()

	buf := &bytes.Buffer{}
	config := &Config{
		Level:  DEBUG,
		Format: "text",
		Output: buf,
	}
	Instance = New(config)

	Info("info message", "key", "value")
	if buf.String() == "" {
		t.Error("expected output from global Info")
	}

	buf.Reset()
	Error("error message", "error", "test")
	if buf.String() == "" {
		t.Error("expected output from global Error")
	}

	buf.Reset()
	Debug("debug message")
	if buf.String() == "" {
		t.Error("expected output from global Debug")
	}

	buf.Reset()
	Warn("warn message")
	if buf.String() == "" {
		t.Error("expected output from global Warn")
	}
}

func TestSetConfig(t *testing.T) {
	original := Instance
	defer func() { Instance = original }()

	newConfig := &Config{
		Level:  WARN,
		Format: "text",
		Output: &bytes.Buffer{},
	}
	SetConfig(newConfig)

	// 验证实例已更新
	if Instance == nil {
		t.Fatal("expected non-nil Instance after SetConfig")
	}

	buf := &bytes.Buffer{}
	newConfig.Output = buf
	SetConfig(newConfig)

	// INFO 应该被过滤
	Instance.Info("test")
	if buf.String() != "" {
		t.Error("expected INFO to be filtered at WARN level")
	}

	// WARN 应该输出
	Instance.Warn("warn msg")
	if buf.String() == "" {
		t.Error("expected WARN to be output")
	}
}
