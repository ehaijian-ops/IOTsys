package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建应用错误
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap 包装错误
func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}

// 预定义错误
var (
	ErrBadRequest     = New("BAD_REQUEST", "请求参数错误", http.StatusBadRequest)
	ErrUnauthorized   = New("UNAUTHORIZED", "未授权访问", http.StatusUnauthorized)
	ErrForbidden      = New("FORBIDDEN", "无权限访问", http.StatusForbidden)
	ErrNotFound       = New("NOT_FOUND", "资源不存在", http.StatusNotFound)
	ErrConflict       = New("CONFLICT", "资源冲突", http.StatusConflict)
	ErrInternalServer = New("INTERNAL_ERROR", "服务器内部错误", http.StatusInternalServerError)
	ErrDeviceOffline  = New("DEVICE_OFFLINE", "设备离线", http.StatusBadRequest)
	ErrCommandTimeout = New("COMMAND_TIMEOUT", "指令执行超时", http.StatusRequestTimeout)
)

// NotFound 资源未找到
func NotFound(resource, id string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s 未找到: %s", resource, id),
		StatusCode: http.StatusNotFound,
	}
}

// ValidationError 校验错误
func ValidationError(msg string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    msg,
		StatusCode: http.StatusUnprocessableEntity,
	}
}
