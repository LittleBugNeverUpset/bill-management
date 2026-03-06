package repo

import (
	"bill-management/internal/model"
	"bill-management/pkg/logger"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserDAO 用户数据访问接口
type UserDAO interface {
	// Create 创建用户
	Create(ctx context.Context, user *model.User) error
	// GetByUsername 根据用户名查询用户（包含软删除用户）
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	// GetByID 根据ID查询用户
	GetByID(ctx context.Context, id int64) (*model.User, error)
	// Update 更新用户信息（非密码字段）
	Update(ctx context.Context, user *model.User) error
	// UpdatePassword 单独更新密码
	UpdatePassword(ctx context.Context, id int64, newPassword string) error
	// List 分页查询用户列表（支持角色、状态筛选）
	List(ctx context.Context, page, pageSize int, role string, status int) ([]*model.User, int64, error)
	// Delete 软删除用户
	Delete(ctx context.Context, id int64) error
	// ExistByUsername 检查用户名是否存在
	ExistByUsername(ctx context.Context, username string) (bool, error)
	// ExistByUserEmail 检查邮箱是否存在
	ExistByUserEmail(ctx context.Context, email string) (bool, error)
}

// userDAO 实现UserDAO接口
type userDAO struct {
	db *gorm.DB // 全局GORM数据库连接
}

// NewUserDAO 创建UserDAO实例
func NewUserDAO(db *gorm.DB) UserDAO {
	return &userDAO{db: db}
}

// Create 创建用户
func (d *userDAO) Create(ctx context.Context, user *model.User) error {
	if err := d.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("用户DAO：创建用户失败", zap.Error(err), zap.String("username", user.Username))
		return err
	}
	logger.Info("用户DAO：创建用户成功", zap.String("username", user.Username))
	return nil
}

// GetByUsername 根据用户名查询用户
func (d *userDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	// Unscoped() 可查询软删除用户（如需仅查未删除，移除该方法）
	err := d.db.WithContext(ctx).Unscoped().Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Debug("用户DAO：用户名不存在", zap.String("username", username))
			return nil, nil // 无数据返回nil，避免上层处理NotFound错误
		}
		logger.Error("用户DAO：查询用户失败（按用户名）", zap.Error(err), zap.String("username", username))
		return nil, err
	}
	return &user, nil
}

// GetByID 根据ID查询用户
func (d *userDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Debug("用户DAO：用户ID不存在", zap.Int64("user_id", id))
			return nil, nil
		}
		logger.Error("用户DAO：查询用户失败（按ID）", zap.Error(err), zap.Int64("user_id", id))
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息（仅更新非密码、非用户名字段）
func (d *userDAO) Update(ctx context.Context, user *model.User) error {
	if user.ID == 0 {
		logger.Warn("用户DAO：更新用户失败，ID为空")
		return errors.New("用户ID不能为空")
	}

	// 只更新指定字段，避免全量更新
	updateData := map[string]interface{}{
		"nickname": user.Nickname,
		"email":    user.Email,
		"role":     user.Role,
		"status":   user.Status,
	}

	err := d.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Updates(updateData).Error
	if err != nil {
		logger.Error("用户DAO：更新用户信息失败", zap.Error(err), zap.Int64("user_id", user.ID))
		return err
	}
	logger.Info("用户DAO：更新用户信息成功", zap.Int64("user_id", user.ID))
	return nil
}

// UpdatePassword 单独更新密码（触发BeforeSave钩子加密）
func (d *userDAO) UpdatePassword(ctx context.Context, id int64, newPassword string) error {
	if id == 0 {
		logger.Warn("用户DAO：更新密码失败，ID为空")
		return errors.New("用户ID不能为空")
	}

	// 构建临时用户对象，仅设置ID和密码（触发BeforeSave钩子）
	user := &model.User{
		ID:       id,
		Password: newPassword,
	}

	// Select指定更新password字段，Omit忽略其他字段
	err := d.db.WithContext(ctx).Model(user).Omit(clause.Associations).Select("password").Updates(user).Error
	if err != nil {
		logger.Error("用户DAO：更新密码失败", zap.Error(err), zap.Int64("user_id", id))
		return err
	}
	logger.Info("用户DAO：更新密码成功", zap.Int64("user_id", id))
	return nil
}

// List 分页查询用户列表
func (d *userDAO) List(ctx context.Context, page, pageSize int, role string, status int) ([]*model.User, int64, error) {
	var (
		users []*model.User
		total int64
	)

	// 基础查询（排除软删除用户）
	query := d.db.WithContext(ctx).Model(&model.User{})

	// 筛选条件
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status >= 0 { // status=-1时不筛选，0/1时筛选
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error("用户DAO：查询用户总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 分页查询（offset=(page-1)*pageSize）
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		logger.Error("用户DAO：分页查询用户失败", zap.Error(err))
		return nil, 0, err
	}

	logger.Info("用户DAO：分页查询用户成功", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.Int64("total", total))
	return users, total, nil
}

// Delete 软删除用户
func (d *userDAO) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		logger.Warn("用户DAO：删除用户失败，ID为空")
		return errors.New("用户ID不能为空")
	}

	// 先检查用户是否存在
	_, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := d.db.WithContext(ctx).Delete(&model.User{}, id).Error; err != nil {
		logger.Error("用户DAO：删除用户失败", zap.Error(err), zap.Int64("user_id", id))
		return err
	}
	logger.Info("用户DAO：删除用户成功", zap.Int64("user_id", id))
	return nil
}

// ExistByUsername 检查用户名是否存在（简化版，供注册校验）
func (d *userDAO) ExistByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		logger.Error("用户DAO：检查用户名是否存在失败", zap.Error(err), zap.String("username", username))
		return false, err
	}
	return count > 0, nil
}

// ExistByUserEmail 检查邮箱是否存在（简化版，供注册校验）
func (d *userDAO) ExistByUserEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		logger.Error("用户DAO：检查邮箱是否存在失败", zap.Error(err), zap.String("email", email))
		return false, err
	}
	return count > 0, nil
}
