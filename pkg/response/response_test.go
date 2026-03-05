package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert" // 可选：简化断言，需安装
)

// 初始化测试环境（禁用 Gin 控制台颜色输出）
func init() {
	gin.SetMode(gin.TestMode)
}

// TestSuccess 测试成功响应
func TestSuccess(t *testing.T) {
	// 1. 创建模拟的 Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 2. 调用 Success 方法
	Success(c, "操作成功")

	// 3. 验证响应结果
	assert.Equal(t, http.StatusOK, w.Code) // 验证 HTTP 状态码
	assert.JSONEq(t, `{
		"code": 200,
		"msg": "操作成功",
		"data": null
	}`, w.Body.String()) // 验证 JSON 结构
}

// TestSuccessWithData 测试带数据的成功响应
func TestSuccessWithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 模拟业务数据
	data := gin.H{
		"id":   1001,
		"name": "测试用户",
	}
	SuccessWithData(c, data, "查询成功")

	// 验证
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{
		"code": 200,
		"msg": "查询成功",
		"data": {"id":1001,"name":"测试用户"}
	}`, w.Body.String())
}

// TestParamError 测试参数错误响应
func TestParamError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ParamError(c, "用户名不能为空")

	// 验证 HTTP 状态码是 400，响应码是 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{
		"code": 400,
		"msg": "用户名不能为空",
		"data": null
	}`, w.Body.String())
}

// TestPageSuccess 测试分页响应
func TestPageSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 模拟分页数据
	data := []gin.H{
		{"id": 1, "name": "用户1"},
		{"id": 2, "name": "用户2"},
	}
	PageSuccess(c, data, 100, 1, 10, "分页查询成功")

	// 验证
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{
		"code": 200,
		"msg": "分页查询成功",
		"data": [{"id":1,"name":"用户1"},{"id":2,"name":"用户2"}],
		"total": 100,
		"page": 1,
		"size": 10
	}`, w.Body.String())
}

// TestAuthError 测试未授权响应
func TestAuthError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AuthError(c, "未登录")

	// 验证 HTTP 状态码是 401，响应码是 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{
		"code": 401,
		"msg": "未登录",
		"data": null
	}`, w.Body.String())
}
