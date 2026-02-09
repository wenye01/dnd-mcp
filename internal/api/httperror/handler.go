// Package httperror 提供HTTP错误处理工具
package httperror

import (
	"errors"
	"net/http"

	"github.com/dnd-mcp/client/internal/api/dto"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ErrorHandler HTTP错误处理器
type ErrorHandler struct{}

// New 创建错误处理器
func New() *ErrorHandler {
	return &ErrorHandler{}
}

// Handle 处理错误并返回适当的HTTP响应
func (h *ErrorHandler) Handle(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 根据错误类型确定HTTP状态码和错误代码
	statusCode, errCode, message := h.classifyError(err)

	// 构建错误响应
	errorResp := dto.NewErrorResponse(errCode, message, nil)

	c.JSON(statusCode, errorResp)
}

// classifyError 分类错误并返回HTTP状态码和错误信息
func (h *ErrorHandler) classifyError(err error) (int, string, string) {
	// 检查预定义错误
	switch {
	case errors.Is(err, errors.ErrSessionNotFound):
		return http.StatusNotFound, "SESSION_NOT_FOUND", "会话不存在"
	case errors.Is(err, errors.ErrMessageNotFound):
		return http.StatusNotFound, "MESSAGE_NOT_FOUND", "消息不存在"
	case errors.Is(err, errors.ErrInvalidArgument):
		return http.StatusBadRequest, "INVALID_ARGUMENT", "无效的请求参数"
	case errors.Is(err, errors.ErrInvalidMaxPlayers):
		return http.StatusBadRequest, "INVALID_MAX_PLAYERS", "玩家数量必须在1-10之间"
	case errors.Is(err, errors.ErrRedisConnection):
		return http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "服务暂时不可用"
	case errors.Is(err, errors.ErrAlreadyExists):
		return http.StatusConflict, "ALREADY_EXISTS", "资源已存在"
	default:
		// 未知错误
		return http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误"
	}
}

// HandleError 全局错误处理函数（便捷方法）
func HandleError(c *gin.Context, err error) {
	New().Handle(c, err)
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("BAD_REQUEST", message, nil)
	c.JSON(http.StatusBadRequest, errorResp)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("UNAUTHORIZED", message, nil)
	c.JSON(http.StatusUnauthorized, errorResp)
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("FORBIDDEN", message, nil)
	c.JSON(http.StatusForbidden, errorResp)
}

// NotFound 返回404错误
func NotFound(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("NOT_FOUND", message, nil)
	c.JSON(http.StatusNotFound, errorResp)
}

// Conflict 返回409错误
func Conflict(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("CONFLICT", message, nil)
	c.JSON(http.StatusConflict, errorResp)
}

// InternalError 返回500错误
func InternalError(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("INTERNAL_ERROR", message, nil)
	c.JSON(http.StatusInternalServerError, errorResp)
}

// ServiceUnavailable 返回503错误
func ServiceUnavailable(c *gin.Context, message string) {
	errorResp := dto.NewErrorResponse("SERVICE_UNAVAILABLE", message, nil)
	c.JSON(http.StatusServiceUnavailable, errorResp)
}

// ValidationError 返回验证错误
func ValidationError(c *gin.Context, details interface{}) {
	errorResp := dto.NewErrorResponse("VALIDATION_ERROR", "请求参数验证失败", details)
	c.JSON(http.StatusBadRequest, errorResp)
}
