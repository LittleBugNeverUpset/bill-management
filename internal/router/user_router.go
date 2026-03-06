// api/router/user_router.go
package router

import (
	"bill-management/internal/controller"
	"bill-management/internal/middleware"
	"bill-management/internal/repository/repo"
	"bill-management/internal/service"
	"bill-management/pkg/jwtutil"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// RegisterUserRouter 注册用户路由
func RegisterUserRouter(r *gin.Engine, db *gorm.DB, jwtService *jwtutil.JWTService) {
	// 初始化依赖
	userRepo := repo.NewUserDAO(db)
	userService := service.NewUserService(userRepo, jwtService)
	userController := controller.NewUserController(userService)

	// 公开接口（无需登录）
	publicGroup := r.Group("/api/user")
	{
		publicGroup.POST("/register", userController.Register)
		publicGroup.POST("/login", userController.Login)
	}

	// 私有接口（需要登录）
	privateGroup := r.Group("/api/user")
	privateGroup.Use(middleware.JWTAuthMiddleware(jwtService)) // JWT鉴权中间件
	{
		privateGroup.POST("/logout", userController.Logout)
	}
}
