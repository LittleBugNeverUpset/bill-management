// api/controller/user_controller.go
package controller

import (
	"bill-management/internal/model"
	"bill-management/internal/service"
	"bill-management/pkg/logger"
	"bill-management/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserController 用户控制器
type UserController struct {
	userService service.UserService // 依赖业务层接口
}

// NewUserController 创建UserController实例
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Register 用户注册接口
// @Summary 用户注册
// @Description 普通用户注册，用户名唯一
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param req body model.UserRegisterRequest true "注册参数"
// @Success 200 {object} gin.H{"code":200,"msg":"注册成功"}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误/用户名已存在"}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误"}
// @Router /api/user/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	// 1. 绑定并校验请求参数
	var req model.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("注册接口：参数绑定失败", zap.Error(err))
		response.Fail(ctx, 400, "参数错误："+err.Error())
		return
	}

	// 2. 调用业务层注册方法
	if err := c.userService.Register(ctx.Request.Context(), &req); err != nil {
		// 业务错误返回400，系统错误返回500
		if err.Error() == "用户名已存在" || err.Error() == "用户名长度需在3-50位之间" || err.Error() == "密码长度需在6-20位之间" {
			response.Fail(ctx, 400, err.Error())
		} else {
			response.Fail(ctx, 500, err.Error())
		}
		return
	}

	// 3. 返回成功响应
	response.Success(ctx, "注册成功")
}

// Login 用户登录接口
// @Summary 用户登录
// @Description 登录成功返回JWT Token，用于后续接口鉴权
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param req body model.UserLoginRequest true "登录参数"
// @Success 200 {object} gin.H{"code":200,"msg":"登录成功","data":{"token":"xxx"}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误"}
// @Failure 401 {object} gin.H{"code":401,"msg":"用户名或密码错误/用户被禁用"}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误"}
// @Router /api/user/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	// 1. 绑定并校验请求参数
	var req model.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("登录接口：参数绑定失败", zap.Error(err))
		response.Fail(ctx, 400, "参数错误："+err.Error())
		return
	}

	// 2. 调用业务层登录方法
	token, err := c.userService.Login(ctx.Request.Context(), &req)
	if err != nil {
		// 业务错误返回401，系统错误返回500
		if err.Error() == "用户名或密码错误" || err.Error() == "用户已被禁用，请联系管理员" || err.Error() == "用户名和密码不能为空" {
			response.Fail(ctx, 401, err.Error())
		} else {
			response.Fail(ctx, 500, err.Error())
		}
		return
	}

	// 3. 返回成功响应（包含Token）
	response.SuccessWithData(ctx, gin.H{"token": token}, "登录成功")
}

// Logout 用户登出接口
// @Summary 用户登出
// @Description 将Token加入黑名单，使其失效
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H{"code":200,"msg":"登出成功"}
// @Failure 401 {object} gin.H{"code":401,"msg":"未登录/Token为空"}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误"}
// @Router /api/user/logout [post]
func (c *UserController) Logout(ctx *gin.Context) {
	// 1. 从Header获取Token（兼容Bearer格式）
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		logger.Warn("登出接口：Authorization头为空")
		response.Fail(ctx, 401, "未登录")
		return
	}

	// 解析Token（去掉Bearer前缀）
	var token string
	parts := []rune(authHeader)
	if len(parts) > 7 && string(parts[:7]) == "Bearer " {
		token = string(parts[7:])
	} else {
		token = authHeader
	}
	if token == "" {
		logger.Warn("登出接口：Token为空")
		response.Fail(ctx, 401, "Token不能为空")
		return
	}

	// 2. 调用业务层登出方法
	if err := c.userService.Logout(ctx.Request.Context(), token); err != nil {
		response.Fail(ctx, 500, err.Error())
		return
	}

	// 3. 返回成功响应
	response.Success(ctx, "登出成功")
}
