// internal/middleware/logger.go
package middleware

import (
	"bill-management/pkg/logger"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 集成zap的Gin日志中间件
// 记录：请求方法、路径、状态码、耗时、客户端IP、请求大小、响应大小等关键信息
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 记录请求开始时间
		startTime := time.Now()

		// 2. 获取请求基础信息
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		// reqHeader := c.Request.Header // 可选：记录请求头（敏感信息需过滤）
		reqSize := c.Request.ContentLength

		// 3. 处理请求（执行后续中间件/处理器）
		c.Next()

		// 4. 请求完成后记录日志
		endTime := time.Now()
		latency := endTime.Sub(startTime) // 耗时
		statusCode := c.Writer.Status()   // 响应状态码
		resSize := c.Writer.Size()        // 响应大小

		// 5. 构造结构化日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.Int64("request_size", reqSize),
			zap.Int("response_size", resSize),
		}

		// 可选：记录请求ID（如果有）
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// 可选：记录错误信息（如果有）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		// 6. 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error(fmt.Sprintf("[%s] %s", method, path), fields...)
		case statusCode >= 400:
			logger.Warn(fmt.Sprintf("[%s] %s", method, path), fields...)
		default:
			logger.Info(fmt.Sprintf("[%s] %s", method, path), fields...)
		}
	}
}

// GinRecovery 异常恢复中间件（结合zap日志记录panic信息）
// 可选补充：防止程序panic崩溃，并记录详细错误日志
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic日志
				logger.Error(
					"请求处理发生panic",
					zap.Any("panic", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Stack("stacktrace"), // 记录堆栈信息
				)

				// 返回500响应
				c.JSON(500, gin.H{
					"code": 500,
					"msg":  "服务器内部错误",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
