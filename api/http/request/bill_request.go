package request

import (
	"bill-management/pkg/logger"
	"bill-management/pkg/responseutil"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateBillRequest 新增账单请求
type CreateBillRequest struct {
	CategoryID uint64  `json:"category_id" binding:"required,gt=0" comment:"分类ID"`
	Amount     float64 `json:"amount" binding:"required" comment:"金额（正数收入/负数支出）"`
	Type       bool    `json:"type" binding:"required" comment:"类型（income/expense）"`
	Remark     string  `json:"remark" binding:"max=255" comment:"备注"`
}

// UpdateBillRequest 更新账单请求
type UpdateBillRequest struct {
	ID         uint64  `json:"id" binding:"required,gt=0" comment:"账单ID"`
	CategoryID uint64  `json:"category_id" binding:"required,gt=0" comment:"分类ID"`
	Amount     float64 `json:"amount" binding:"required" comment:"金额（正数收入/负数支出）"`
	Type       bool    `json:"type" binding:"required" comment:"类型（income/expense）"`
	Remark     string  `json:"remark" binding:"max=255" comment:"备注"`
}

// ListBillRequest 账单列表查询请求
type ListBillRequest struct {
	CategoryID uint64 `form:"category_id" comment:"分类ID（0表示不筛选）"`
	Type       bool   `form:"type" binding:"omitempty" comment:"类型（income/expense）"`
	StartTime  string `form:"start_time" comment:"开始时间（格式：2024-01-01）"`
	EndTime    string `form:"end_time" comment:"结束时间（格式：2024-01-31）"`
	Page       int    `form:"page" binding:"required,gt=0" comment:"页码"`
	PageSize   int    `form:"page_size" binding:"required,gt=0,lte=100" comment:"每页条数（最大100）"`
}

// BatchDeleteBillRequest 批量删除账单请求
type BatchDeleteBillRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1" comment:"账单ID列表"`
}

// ParseTime 解析时间字符串为time.Time
func (r *ListBillRequest) ParseTime() (startTime, endTime time.Time, err error) {
	if r.StartTime != "" {
		startTime, err = time.Parse("2006-01-02", r.StartTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if r.EndTime != "" {
		endTime, err = time.Parse("2006-01-02", r.EndTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// 结束时间补全为当天23:59:59
		endTime = endTime.Add(24*time.Hour - time.Second)
	}
	return
}

// GetUserID 从Gin Context获取用户ID（JWT中间件存入）
func GetUserID(c *gin.Context) uint64 {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		logger.Error("从Context获取user_id失败：未找到该key")
		responseutil.AuthError(c, "用户未登录")
		return 0
	}
	// 兼容多种数字类型（int/int64/uint64），避免类型问题
	var userIDInt int
	switch v := userIDVal.(type) {
	case int:
		userIDInt = v
	case int64:
		userIDInt = int(v)
	case uint64:
		userIDInt = int(v)
	default:
		logger.Error("user_id类型错误", zap.Any("userIDVal", userIDVal), zap.Any("type", fmt.Sprintf("%T", userIDVal)))
		responseutil.ServerError(c, "用户ID格式错误")
		return 0
	}

	return uint64(userIDInt)
}

// 	// 关键修复：先断言为int（JWT存入的是int），再转为uint64
// 	userIDInt, ok := userIDVal.(int)
// 	if !ok {
// 		logger.Error("user_id类型错误", zap.Any("userIDVal", userIDVal))
// 		responseutil.ServerError(c, "用户ID格式错误")
// 		return 0
// 	}

// 	// 转为uint64（适配数据库模型的UserID类型）
// 	return uint64(userIDInt)
// }
