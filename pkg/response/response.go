package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义响应码常量
const (
	CodeSuccess  = 200 // 成功
	CodeError    = 500 // 服务器内部错误
	CodeParam    = 400 // 参数错误
	CodeAuth     = 401 // 未授权
	CodeForbid   = 403 // 禁止访问
	CodeNotFound = 404 // 资源不存在
)

// BaseResponse 基础响应结构
type BaseResponse struct {
	Code int         `json:"code"` // 响应码
	Msg  string      `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 业务数据（可选）
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code  int         `json:"code"`  // 响应码
	Msg   string      `json:"msg"`   // 提示信息
	Data  interface{} `json:"data"`  // 分页数据列表
	Total int64       `json:"total"` // 总条数
	Page  int         `json:"page"`  // 当前页码
	Size  int         `json:"size"`  // 每页条数
}

// ========== 通用响应方法 ==========

// Success 成功响应（无数据）
func Success(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, BaseResponse{
		Code: CodeSuccess,
		Msg:  msg,
		Data: nil,
	})
}

// SuccessWithData 成功响应（带数据）
func SuccessWithData(c *gin.Context, data interface{}, msg ...string) {
	msgStr := "操作成功"
	if len(msg) > 0 {
		msgStr = msg[0]
	}
	c.JSON(http.StatusOK, BaseResponse{
		Code: CodeSuccess,
		Msg:  msgStr,
		Data: data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	// 确保 HTTP 状态码合理
	httpCode := http.StatusInternalServerError
	switch code {
	case CodeParam:
		httpCode = http.StatusBadRequest
	case CodeAuth:
		httpCode = http.StatusUnauthorized
	case CodeForbid:
		httpCode = http.StatusForbidden
	case CodeNotFound:
		httpCode = http.StatusNotFound
	}
	c.JSON(httpCode, BaseResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// FailWithData 失败响应（带数据）
func FailWithData(c *gin.Context, code int, msg string, data interface{}) {
	httpCode := http.StatusInternalServerError
	switch code {
	case CodeParam:
		httpCode = http.StatusBadRequest
	case CodeAuth:
		httpCode = http.StatusUnauthorized
	case CodeForbid:
		httpCode = http.StatusForbidden
	case CodeNotFound:
		httpCode = http.StatusNotFound
	}
	c.JSON(httpCode, BaseResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, data interface{}, total int64, page, size int, msg ...string) {
	msgStr := "查询成功"
	if len(msg) > 0 {
		msgStr = msg[0]
	}
	c.JSON(http.StatusOK, PageResponse{
		Code:  CodeSuccess,
		Msg:   msgStr,
		Data:  data,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// ========== 快捷方法（简化调用） ==========

// ParamError 参数错误响应
func ParamError(c *gin.Context, msg string) {
	Fail(c, CodeParam, msg)
}

// AuthError 未授权响应
func AuthError(c *gin.Context, msg string) {
	Fail(c, CodeAuth, msg)
}

// NotFoundError 资源不存在响应
func NotFoundError(c *gin.Context, msg string) {
	Fail(c, CodeNotFound, msg)
}

// ServerError 服务器内部错误响应
func ServerError(c *gin.Context, msg string) {
	Fail(c, CodeError, msg)
}
