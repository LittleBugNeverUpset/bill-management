// pkg/model/user.go
package model

import (
	"time"

	"bill-management/pkg/crypto"

	"gorm.io/gorm"
)

// User 用户实体
type User struct {
	ID       int64  `gorm:"primarykey;autoIncrement;primaryKey;" json:"id"` // 用户ID
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"`   // 用户名（唯一）
	Password string `gorm:"size:100;not null" json:"-"`                     // 密码（加密存储，返回时隐藏）
	Nickname string `gorm:"size:50;default:''" json:"nickname"`             // 昵称
	Email    string `gorm:"size:100;uniqueIndex;default:''" json:"email"`   // 邮箱（唯一）
	Role     string `gorm:"size:20;default:'user'" json:"role"`             // 角色：admin/manager/user
	Status   int    `gorm:"default:1" json:"status"`                        // 状态：1-正常 0-禁用

	CreatedAt time.Time      `json:"created_at"`     // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`     // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定数据库表名
func (u *User) TableName() string {
	return "users"
}

// BeforeSave GORM钩子：保存前加密密码
func (u *User) BeforeSave(tx *gorm.DB) error {
	// 仅当密码有修改时加密（避免重复加密）
	if len(u.Password) < 60 {
		hashedPwd, err := crypto.BcryptEncrypt(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPwd
	}
	return nil
}

// CheckPassword 验证密码是否正确
func (u *User) CheckPassword(password string) bool {
	return crypto.BcryptVerify(password, u.Password)
}

// UserLoginRequest 登录请求参数
type UserLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=20"` // 密码
}

// UserRegisterRequest 注册请求参数
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=20"` // 密码
	Nickname string `json:"nickname" binding:"max=50"`                // 昵称
	Email    string `json:"email" binding:"email,max=100"`            // 邮箱
}

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`               // 昵称
	Email    string `json:"email" binding:"email,max=100"`           // 邮箱
	Role     string `json:"role" binding:"oneof=admin manager user"` // 角色（仅管理员可修改）
	Status   int    `json:"status" binding:"oneof=0 1"`              // 状态（仅管理员可修改）
}

// UserChangePwdRequest 密码修改请求
type UserChangePwdRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=20"` // 原密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"` // 新密码
}
