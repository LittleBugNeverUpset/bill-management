package router

import (
	"bill-management/internal/controller"
	"bill-management/internal/middleware"
	"bill-management/internal/repository/redis_repo"
	"bill-management/internal/repository/repo"
	"bill-management/internal/service"
	"bill-management/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterBillRouter 注册账单模块路由
func RegisterBillRouter(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, jwtService *jwtutil.JWTService) {
	billRepo := repo.NewBillDAO(db)
	billCache := redis_repo.NewBillCache(redisClient)
	billService := service.NewBillService(billRepo, billCache, jwtService)
	billController := controller.NewBillController(billService)
	// 账单路由组（需要JWT鉴权）
	billGroup := r.Group("/bill")
	billGroup.Use(middleware.JWTAuthMiddleware(jwtService))
	{
		billGroup.POST("", billController.CreateBill)              // 新增账单
		billGroup.GET("", billController.ListBill)                 // 查询账单列表
		billGroup.GET("/:id", billController.GetBillByID)          // 查询账单详情
		billGroup.PUT("", billController.UpdateBill)               // 更新账单
		billGroup.DELETE("/:id", billController.DeleteBill)        // 删除账单
		billGroup.DELETE("/batch", billController.BatchDeleteBill) // 批量删除账单
	}
}
