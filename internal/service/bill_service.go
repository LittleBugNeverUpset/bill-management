package service

import (
	"context"
	"errors"
	"time"

	"bill-management/api/http/request"
	"bill-management/api/http/response"
	"bill-management/internal/model"
	"bill-management/internal/repository/redis_repo"
	"bill-management/internal/repository/repo"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"

	"go.uber.org/zap"
)

// BillService 账单业务逻辑接口
type BillService interface {
	// CreateBill 新增账单
	CreateBill(ctx context.Context, req *request.CreateBillRequest) (uint64, error)
	// GetBillByID 查询账单详情
	GetBillByID(ctx context.Context, userID, billID uint64) (*response.BillResponse, error)
	// ListBill 查询账单列表
	ListBill(ctx context.Context, userID uint64, req *request.ListBillRequest) (*response.BillListResponse, error)
	// UpdateBill 更新账单
	UpdateBill(ctx context.Context, userID uint64, req *request.UpdateBillRequest) error
	// DeleteBill 删除账单
	DeleteBill(ctx context.Context, userID, billID uint64) error
	// BatchDeleteBill 批量删除账单
	BatchDeleteBill(ctx context.Context, userID uint64, req *request.BatchDeleteBillRequest) error
}

// billService 实现BillService接口
type billService struct {
	billDAO    repo.BillDAO
	jwtService *jwtutil.JWTService
	billCache  redis_repo.BillCache
}

// NewBillService 创建账单业务逻辑实例
func NewBillService(billDAO repo.BillDAO, billCache redis_repo.BillCache, jwtService *jwtutil.JWTService) BillService {
	return &billService{
		billDAO:    billDAO,
		jwtService: jwtService, // JWT工具（生成/解析/拉黑Token）
		billCache:  billCache,
	}
}

// CreateBill 新增账单
func (s *billService) CreateBill(ctx context.Context, req *request.CreateBillRequest) (uint64, error) {

	userID := ctx.Value("user_id").(uint64)
	logger.Infof("UserId %d", userID)
	// 2. 构建账单模型
	bill := &model.Bill{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Type:       req.Type,
		Remark:     req.Remark,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 3. 新增账单
	logger.Info("开始创建新订单")
	err := s.billDAO.Create(ctx, bill)
	if err != nil {
		logger.Error("新增账单失败", zap.Uint64("userID", userID), zap.Error(err))
		return 0, err
	}

	// 4. 清理该用户当月的统计缓存（账单新增会影响统计）
	month := bill.CreatedAt.Format("2006-01")
	err = s.billCache.DelBillStat(ctx, userID, month)
	if err != nil {
		logger.Warn("清理账单统计缓存失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
	}

	return bill.ID, nil
}

// GetBillByID 查询账单详情
func (s *billService) GetBillByID(ctx context.Context, userID, billID uint64) (*response.BillResponse, error) {
	bill, err := s.billDAO.GetByID(ctx, userID, billID)
	if err != nil {
		logger.Error("查询账单详情失败", zap.Uint64("userID", userID), zap.Uint64("billID", billID), zap.Error(err))
		return nil, err
	}
	if bill == nil {
		return nil, nil
	}
	return &response.BillResponse{
		ID:         bill.ID,
		UserID:     bill.UserID,
		CategoryID: bill.CategoryID,
		Amount:     bill.Amount,
		Type:       bill.Type,
		Remark:     bill.Remark,
		CreatedAt:  bill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  bill.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ListBill 查询账单列表
func (s *billService) ListBill(ctx context.Context, userID uint64, req *request.ListBillRequest) (*response.BillListResponse, error) {
	// 1. 解析时间
	startTime, endTime, err := req.ParseTime()
	if err != nil {
		logger.Error("解析时间失败", zap.Uint64("userID", userID), zap.Error(err))
		return nil, errors.New("时间格式错误，正确格式：2024-01-01")
	}

	// 2. 构建查询参数
	repoReq := repo.BillListRequest{
		CategoryID: req.CategoryID,
		Type:       req.Type,
		StartTime:  startTime,
		EndTime:    endTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 3. 查询账单列表
	bills, total, err := s.billDAO.List(ctx, userID, repoReq)
	if err != nil {
		logger.Error("查询账单列表失败", zap.Uint64("userID", userID), zap.Error(err))
		return nil, err
	}

	// 4. 转换为响应体
	listResp := response.ConvertBillListToResponse(bills, total, req.Page, req.PageSize)
	return &listResp, nil
}

// UpdateBill 更新账单
func (s *billService) UpdateBill(ctx context.Context, userID uint64, req *request.UpdateBillRequest) error {
	// 1. 校验账单是否存在且属于当前用户
	bill, err := s.billDAO.GetByID(ctx, userID, req.ID)
	if err != nil {
		logger.Error("查询账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", req.ID), zap.Error(err))
		return err
	}
	if bill == nil {
		return errors.New("账单不存在或无权限")
	}

	// 3. 更新账单信息
	bill.CategoryID = req.CategoryID
	bill.Amount = req.Amount
	bill.Type = req.Type
	bill.Remark = req.Remark
	bill.UpdatedAt = time.Now()

	err = s.billDAO.Update(ctx, bill)
	if err != nil {
		logger.Error("更新账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", req.ID), zap.Error(err))
		return err
	}

	// 4. 清理该用户账单所属月份的统计缓存
	month := bill.CreatedAt.Format("2006-01")
	err = s.billCache.DelBillStat(ctx, userID, month)
	if err != nil {
		logger.Warn("清理账单统计缓存失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
	}

	return nil
}

// DeleteBill 删除账单
func (s *billService) DeleteBill(ctx context.Context, userID, billID uint64) error {
	// 1. 校验账单是否存在且属于当前用户
	bill, err := s.billDAO.GetByID(ctx, userID, billID)
	if err != nil {
		logger.Error("查询账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", billID), zap.Error(err))
		return err
	}
	if bill == nil {
		return errors.New("账单不存在或无权限")
	}

	// 2. 删除账单
	err = s.billDAO.Delete(ctx, userID, billID)
	if err != nil {
		logger.Error("删除账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", billID), zap.Error(err))
		return err
	}

	// 3. 清理该用户账单所属月份的统计缓存
	month := bill.CreatedAt.Format("2006-01")
	err = s.billCache.DelBillStat(ctx, userID, month)
	if err != nil {
		logger.Warn("清理账单统计缓存失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
	}

	return nil
}

// BatchDeleteBill 批量删除账单
func (s *billService) BatchDeleteBill(ctx context.Context, userID uint64, req *request.BatchDeleteBillRequest) error {
	// 1. 批量删除账单
	err := s.billDAO.BatchDelete(ctx, userID, req.IDs)
	if err != nil {
		logger.Error("批量删除账单失败", zap.Uint64("userID", userID), zap.Error(err))
		return err
	}

	// 2. 清理该用户所有账单统计缓存（批量删除无法精准判断月份，直接清理全量）
	err = s.billCache.DelUserAllBillStat(ctx, userID)
	if err != nil {
		logger.Warn("清理用户账单统计缓存失败", zap.Uint64("userID", userID), zap.Error(err))
	}

	return nil
}
