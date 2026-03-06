package util_test

import (
	"bill-management/internal/model"
	"bill-management/internal/repository/repo"
	"bill-management/pkg/databaseutil"
	"bill-management/pkg/logger"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 全局测试变量
var (
	testDB   *gorm.DB
	userRepo repo.UserDAO
	testCtx  = context.Background()
	// 测试用用户数据
	testUser = &model.User{
		Username: "test_user_001",
		Password: "123456", // 实际会被BeforeSave钩子加密
		Nickname: "测试用户",
		Email:    "test@example.com",
		Role:     "user",
		Status:   1,
	}
)

func setupTestPostgreSQL(t *testing.T) {
	// 初始化测试数据库连接
	testDB = databaseutil.InitPostgreSQL()
	if testDB == nil {
		t.Fatal("初始化测试数据库连接失败")
	}

	// 初始化UserRepo实例
	userRepo = repo.NewUserDAO(testDB)
}

// TestUserRepo_Create 测试：注册用户 → MySQL能查到数据
func TestUserRepo_Create(t *testing.T) {
	setupTestPostgreSQL(t)
	// 前置：先删除可能存在的同名测试用户（避免冲突）
	_ = testDB.Where("username = ?", testUser.Username).Delete(&model.User{}).Error

	// 步骤1：执行创建用户
	err := userRepo.Create(testCtx, testUser)
	assert.NoError(t, err, "创建用户失败")

	// 步骤2：从数据库查询，验证数据存在
	var dbUser model.User
	err = testDB.Where("username = ?", testUser.Username).First(&dbUser).Error
	assert.NoError(t, err, "查询创建的用户失败")

	// 断言关键字段正确
	assert.Equal(t, testUser.Username, dbUser.Username, "用户名不一致")
	assert.Equal(t, testUser.Nickname, dbUser.Nickname, "昵称不一致")
	assert.Equal(t, testUser.Email, dbUser.Email, "邮箱不一致")
	assert.Equal(t, testUser.Role, dbUser.Role, "角色不一致")
	assert.Equal(t, testUser.Status, dbUser.Status, "状态不一致")
	logger.Infof("db_user: %s", dbUser.Password)
	logger.Infof("test_user: %s", testUser.Password)
	// assert.NotEqual(t, testUser.Password, dbUser.Password, "密码未加密（应该被bcrypt加密）")

	// 保存用户ID，供后续测试使用
	testUser.ID = dbUser.ID
	logger.Infof("测试用户ID: %d", testUser.ID)
}

// TestUserRepo_GetByUsername 测试：按用户名查询 → 返回正确用户
func TestUserRepo_GetByUsername(t *testing.T) {
	setupTestPostgreSQL(t)
	logger.Infof("测试用户ID: %s", testUser)
	// 前置：确保测试用户已存在（依赖上一个测试的创建结果）
	// if testUser.ID == 0 {
	// 	t.Skip("测试用户未创建，跳过该测试")
	// }

	// 步骤1：按用户名查询
	user, err := userRepo.GetByUsername(testCtx, testUser.Username)
	assert.NoError(t, err, "按用户名查询失败")
	assert.NotNil(t, user, "未查询到用户")

	// 断言查询结果正确
	// assert.Equal(t, testUser.ID, user.ID, "用户ID不一致")
	assert.Equal(t, testUser.Username, user.Username, "用户名不一致")
	assert.Equal(t, testUser.Nickname, user.Nickname, "昵称不一致")
}

// TestUserRepo_Update 测试：更新用户信息 → MySQL数据同步更新
func TestUserRepo_Update(t *testing.T) {
	setupTestPostgreSQL(t)
	// 前置：确保测试用户已存在
	// if testUser.ID == 0 {
	// 	t.Skip("测试用户未创建，跳过该测试")
	// }

	// 步骤1：构造更新数据
	updateUser := &model.User{
		// ID:       testUser.ID,
		Nickname: "更新后的测试用户",                // 新昵称
		Email:    "update_test@example.com", // 新邮箱
		Role:     "manager",                 // 新角色
		Status:   0,                         // 新状态（禁用）
	}

	// 步骤2：执行更新
	err := userRepo.Update(testCtx, updateUser)
	assert.NoError(t, err, "更新用户信息失败")

	// 步骤3：从数据库查询，验证更新结果
	var dbUser model.User
	err = testDB.Where("id = ?", testUser.ID).First(&dbUser).Error
	assert.NoError(t, err, "查询更新后的用户失败")

	// 断言更新字段生效
	assert.Equal(t, updateUser.Nickname, dbUser.Nickname, "昵称未更新")
	assert.Equal(t, updateUser.Email, dbUser.Email, "邮箱未更新")
	assert.Equal(t, updateUser.Role, dbUser.Role, "角色未更新")
	assert.Equal(t, updateUser.Status, dbUser.Status, "状态未更新")

	// 断言未更新字段不变（用户名、密码）
	assert.Equal(t, testUser.Username, dbUser.Username, "用户名被错误更新")
	assert.NotEmpty(t, dbUser.Password, "密码被清空")
}

// 可选：测试删除用户（补充测试）
func TestUserRepo_Delete(t *testing.T) {
	setupTestPostgreSQL(t)
	// 前置：确保测试用户已存在
	// if testUser.ID == 0 {
	// 	t.Skip("测试用户未创建，跳过该测试")
	// }

	// 步骤1：执行删除
	err := userRepo.Delete(testCtx, 16)
	assert.NoError(t, err, "删除用户失败")

	// 步骤2：验证用户已被软删除
	var dbUser model.User
	err = testDB.Unscoped().Where("id = ?", testUser.ID).First(&dbUser).Error
	assert.NoError(t, err, "查询软删除用户失败")
	assert.NotZero(t, dbUser.DeletedAt.Time, "用户未被软删除")

	// 步骤3：普通查询（不包含软删除）应返回nil
	user, err := userRepo.GetByID(testCtx, testUser.ID)
	assert.NoError(t, err)
	assert.Nil(t, user, "软删除用户仍能被查询到")
}
