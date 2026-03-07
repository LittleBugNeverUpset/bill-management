package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"bill-management/api/http/request"
	"bill-management/internal/service"
	"bill-management/pkg/logger"
	"bill-management/pkg/responseutil"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BillController 账单控制器
type BillController struct {
	billService service.BillService
}

// NewBillController 创建账单控制器实例
func NewBillController(billService service.BillService) *BillController {
	return &BillController{
		billService: billService,
	}
}

// CreateBill 新增账单
// @Summary 新增账单
// @Description 新增用户账单
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateBillRequest true "新增账单参数"
// @Success 200 {object} response.Response{data=uint64} "成功：返回账单ID"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill [post]
func (c *BillController) CreateBill(ctx *gin.Context) {

	userID := request.GetUserID(ctx)
	// 补充日志：打印获取到的userID，确认JWT解析正常
	logger.Info("开始新增账单", zap.Uint64("userID", userID))

	// 2. 绑定请求参数
	var req request.CreateBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 补充详细错误日志
		logger.Error("绑定新增账单参数失败",
			zap.Uint64("userID", userID),
			zap.Error(err),
			zap.Any("request_body", ctx.Request.Body), // 打印请求体
		)
		responseutil.Fail(ctx, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	// 3. 调用Service
	billID, err := c.billService.CreateBill(ctx, &req)
	if err != nil {
		// 补充详细错误日志（关键：打印Service层的具体错误）
		logger.Error("新增账单Service层执行失败",
			zap.Uint64("userID", userID),
			zap.Any("req", req), // 打印请求参数
			zap.Error(err),      // 打印具体错误信息
		)
		responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. 返回成功响应
	logger.Info("新增账单成功", zap.Uint64("userID", userID), zap.Uint64("billID", billID))
	responseutil.Success(ctx, fmt.Sprint(billID))
}

// GetBillByID 查询账单详情
// @Summary 查询账单详情
// @Description 根据ID查询用户账单详情
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path uint64 true "账单ID"
// @Success 200 {object} response.Response{data=response.BillResponse} "成功：返回账单详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "账单不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill/{id} [get]
func (c *BillController) GetBillByID(ctx *gin.Context) {
	// 1. 获取用户ID和账单ID
	userID := request.GetUserID(ctx)
	billID := ctx.Param("id")
	billIDUint, err := strconv.ParseUint(billID, 10, 64)
	if err != nil {
		logger.Error("解析账单ID失败", zap.String("billID", billID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusBadRequest, "账单ID格式错误")
		return
	}

	// 2. 调用Service
	billResp, err := c.billService.GetBillByID(ctx, userID, billIDUint)
	if err != nil {
		logger.Error("查询账单详情失败", zap.Uint64("userID", userID), zap.Uint64("billID", billIDUint), zap.Error(err))
		responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if billResp == nil {
		responseutil.Fail(ctx, http.StatusNotFound, "账单不存在")
		return
	}

	// 3. 返回成功响应
	responseutil.Success(ctx, fmt.Sprintf("账单ID:%d", billIDUint))
}

// ListBill 查询账单列表
// @Summary 查询账单列表
// @Description 分页查询用户账单列表，支持按分类、类型、时间范围筛选
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param category_id query uint64 false "分类ID"
// @Param type query string false "类型（income/expense）"
// @Param start_time query string false "开始时间（格式：2024-01-01）"
// @Param end_time query string false "结束时间（格式：2024-01-31）"
// @Param page query int true "页码"
// @Param page_size query int true "每页条数"
// @Success 200 {object} response.Response{data=response.BillListResponse} "成功：返回账单列表"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill [get]
func (c *BillController) ListBill(ctx *gin.Context) {
	// 1. 获取用户ID
	userID := request.GetUserID(ctx)

	// 2. 绑定请求参数
	var req request.ListBillRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logger.Error("绑定账单列表参数失败", zap.Uint64("userID", userID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	// 3. 调用Service
	listResp, err := c.billService.ListBill(ctx, userID, &req)
	if err != nil {
		logger.Error("查询账单列表失败", zap.Uint64("userID", userID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. 返回成功响应
	responseutil.SuccessWithData(ctx, listResp)
}

// UpdateBill 更新账单
// @Summary 更新账单
// @Description 更新用户账单信息
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.UpdateBillRequest true "更新账单参数"
// @Success 200 {object} response.Response "成功：返回ok"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "账单不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill [put]
func (c *BillController) UpdateBill(ctx *gin.Context) {
	// 1. 获取用户ID
	userID := request.GetUserID(ctx)

	// 2. 绑定请求参数
	var req request.UpdateBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定更新账单参数失败", zap.Uint64("userID", userID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	// 3. 调用Service
	err := c.billService.UpdateBill(ctx, userID, &req)
	if err != nil {
		logger.Error("更新账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", req.ID), zap.Error(err))
		if err.Error() == "账单不存在或无权限" {
			responseutil.Fail(ctx, http.StatusNotFound, err.Error())
		} else {
			responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// 4. 返回成功响应
	responseutil.Success(ctx, "更新成功")
}

// DeleteBill 删除账单
// @Summary 删除账单
// @Description 删除用户账单
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path uint64 true "账单ID"
// @Success 200 {object} response.Response "成功：返回ok"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "账单不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill/{id} [delete]
func (c *BillController) DeleteBill(ctx *gin.Context) {
	// 1. 获取用户ID和账单ID
	userID := request.GetUserID(ctx)
	billID := ctx.Param("id")
	billIDUint, err := strconv.ParseUint(billID, 10, 64)
	if err != nil {
		logger.Error("解析账单ID失败", zap.String("billID", billID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusBadRequest, "账单ID格式错误")
		return
	}

	// 2. 调用Service
	err = c.billService.DeleteBill(ctx, userID, billIDUint)
	if err != nil {
		logger.Error("删除账单失败", zap.Uint64("userID", userID), zap.Uint64("billID", billIDUint), zap.Error(err))
		if err.Error() == "账单不存在或无权限" {
			responseutil.Fail(ctx, http.StatusNotFound, err.Error())
		} else {
			responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// 3. 返回成功响应
	responseutil.Success(ctx, "删除成功")
}

// BatchDeleteBill 批量删除账单
// @Summary 批量删除账单
// @Description 批量删除用户账单
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.BatchDeleteBillRequest true "批量删除账单参数"
// @Success 200 {object} response.Response "成功：返回ok"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/bill/batch [delete]
func (c *BillController) BatchDeleteBill(ctx *gin.Context) {
	// 1. 获取用户ID
	userID := request.GetUserID(ctx)

	// 2. 绑定请求参数
	var req request.BatchDeleteBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定批量删除账单参数失败", zap.Uint64("userID", userID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	// 3. 调用Service
	err := c.billService.BatchDeleteBill(ctx, userID, &req)
	if err != nil {
		logger.Error("批量删除账单失败", zap.Uint64("userID", userID), zap.Error(err))
		responseutil.Fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. 返回成功响应
	responseutil.Success(ctx, "批量删除成功")
}
