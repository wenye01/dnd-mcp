// Package errors 提供应用级别的错误定义
package errors

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	// ErrSessionNotFound 会话不存在
	ErrSessionNotFound = errors.New("session not found")

	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = errors.New("message not found")

	// ErrInvalidArgument 无效参数
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrRedisConnection Redis 连接失败
	ErrRedisConnection = errors.New("redis connection failed")

	// ErrAlreadyExists 已存在
	ErrAlreadyExists = errors.New("already exists")
)

// Is 检查错误是否是目标类型
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Wrap 包装错误,添加上下文信息
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf 包装错误,添加格式化的上下文信息
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
