package repo

import (
	"context"
	"errors"
	"time"

	"bill-management/internal/model"
	"bill-management/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BillRepo 账单数据访问接口
type BillDAO interface {
	// Create 新增账单
	Create(ctx context.Context, bill *model.Bill) error
	// GetByID 按ID查询账单（仅所属用户）
	GetByID(ctx context.Context, userID, billID uint64) (*model.Bill, error)
	// List 筛选账单（用户ID+时间范围+分类ID+类型）
	List(ctx context.Context, userID uint64, req BillListRequest) ([]*model.Bill, int64, error)
	// Update 更新账单（仅所属用户）
	Update(ctx context.Context, bill *model.Bill) error
	// Delete 软删除账单（仅所属用户）
	Delete(ctx context.Context, userID, billID uint64) error
	// BatchDelete 批量删除账单（仅所属用户）
	BatchDelete(ctx context.Context, userID uint64, billIDs []uint64) error
}

// BillListRequest 账单列表查询参数
type BillListRequest struct {
	CategoryID uint64    // 分类ID（0表示不筛选）
	Type       bool      // 类型（income/expense，空表示不筛选）
	StartTime  time.Time // 开始时间（零值表示不筛选）
	EndTime    time.Time // 结束时间（零值表示不筛选）
	Page       int       // 页码
	PageSize   int       // 每页条数
}

// billDAO 实现BillRepo接口
type billDAO struct {
	db *gorm.DB
}

// NewBillRepo 创建账单数据访问实例
func NewBillDAO(db *gorm.DB) BillDAO {
	return &billDAO{db: db}
}

// Create 新增账单
func (r *billDAO) Create(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

// GetByID 按ID查询账单
func (r *billDAO) GetByID(ctx context.Context, userID, billID uint64) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", billID, userID).First(&bill).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 无数据返回nil，避免上层处理error
		}

		return nil, err
	}
	return &bill, nil
}

// List 筛选账单
func (r *billDAO) List(ctx context.Context, userID uint64, req BillListRequest) ([]*model.Bill, int64, error) {
	// 构建查询条件
	query := r.db.WithContext(ctx).Model(&model.Bill{}).Where("user_id = ?", userID)

	// 分类筛选
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}

	// // 类型筛选
	// if req.Type != "" {
	// 	query = query.Where("type = ?", req.Type)
	// }

	// 时间范围筛选
	if !req.StartTime.IsZero() {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		// logger.Fatalf("查询账单总数失败,userID %d.   %_+v", userID, err)
		logger.Error("查询账单总数失败", zap.Error(err), zap.Error(err), zap.Int64("UserI", int64(userID)))
		return nil, 0, err
	}

	// 分页查询
	var bills []*model.Bill
	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&bills).Error
	if err != nil {
		// logger.Fatalf("分页查询账单失败,userID %d.   %_+v", userID, err)
		logger.Error("分页查询账单失败", zap.Error(err), zap.Error(err), zap.Int64("UserI", int64(userID)))
		return nil, 0, err
	}

	return bills, total, nil
}

// Update 更新账单
func (r *billDAO) Update(ctx context.Context, bill *model.Bill) error {
	// 仅更新允许修改的字段：category_id、amount、type、remark、updated_at
	return r.db.WithContext(ctx).Model(bill).Where("user_id = ?", bill.UserID).Updates(map[string]interface{}{
		"category_id": bill.CategoryID,
		"amount":      bill.Amount,
		"type":        bill.Type,
		"remark":      bill.Remark,
		"updated_at":  time.Now(),
	}).Error
}

// Delete 软删除账单
func (r *billDAO) Delete(ctx context.Context, userID, billID uint64) error {
	return r.db.WithContext(ctx).Model(&model.Bill{}).Where("id = ? AND user_id = ?", billID, userID).Update("deleted_at", time.Now()).Error
}

// BatchDelete 批量删除账单
func (r *billDAO) BatchDelete(ctx context.Context, userID uint64, billIDs []uint64) error {
	if len(billIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.Bill{}).Where("id IN (?) AND user_id = ?", billIDs, userID).Update("deleted_at", time.Now()).Error
}
