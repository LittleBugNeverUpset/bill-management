package response

import (
	"bill-management/internal/model"
)

// BillResponse 账单响应体
type BillResponse struct {
	ID         uint64  `json:"id"`          // 账单ID
	UserID     uint64  `json:"user_id"`     // 所属用户ID
	CategoryID uint64  `json:"category_id"` // 分类ID
	Amount     float64 `json:"amount"`      // 金额
	Type       bool    `json:"type"`        // 类型（（income(1)/expense(0)）
	Remark     string  `json:"remark"`      // 备注
	CreatedAt  string  `json:"created_at"`  // 创建时间（格式化）
	UpdatedAt  string  `json:"updated_at"`  // 更新时间（格式化）
}

// BillListResponse 账单列表响应体
type BillListResponse struct {
	List  []BillResponse `json:"list"`  // 账单列表
	Total int64          `json:"total"` // 总数
	Page  int            `json:"page"`  // 当前页码
	Size  int            `json:"size"`  // 每页条数
}

// ConvertBillToResponse 转换模型为响应体
func ConvertBillToResponse(bill *model.Bill) BillResponse {
	return BillResponse{
		ID:         bill.ID,
		UserID:     bill.UserID,
		CategoryID: bill.CategoryID,
		Amount:     bill.Amount,
		Type:       bill.Type,
		Remark:     bill.Remark,
		CreatedAt:  bill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  bill.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

}

// ConvertBillListToResponse 转换账单列表为响应体
func ConvertBillListToResponse(bills []*model.Bill, total int64, page, pageSize int) BillListResponse {
	list := make([]BillResponse, 0, len(bills))
	for _, bill := range bills {
		list = append(list, ConvertBillToResponse(bill))
	}
	return BillListResponse{
		List:  list,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}
}
