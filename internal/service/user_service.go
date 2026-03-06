// pkg/service/user_service.go
package service

import (
	"bill-management/internal/model"
	"bill-management/internal/repository/repo"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"
	"context"
	"errors"

	"go.uber.org/zap"
)

// UserService 用户业务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *model.UserRegisterRequest) error
	// Login 用户登录（返回JWT Token）
	Login(ctx context.Context, req *model.UserLoginRequest) (string, error)
	// Logout 用户登出（Token加入黑名单）
	Logout(ctx context.Context, token string) error
}

// userService 实现UserService接口
type userService struct {
	userRepo   repo.UserDAO        // 数据访问层
	jwtService *jwtutil.JWTService // JWT工具（生成/解析/拉黑Token）
}

// NewUserService 创建UserService实例
func NewUserService(userRepo repo.UserDAO, jwtService *jwtutil.JWTService) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *model.UserRegisterRequest) error {
	// 1. 业务规则校验：用户名/密码长度
	if len(req.Username) < 3 || len(req.Username) > 50 {
		logger.Warn("注册失败：用户名长度不符合要求", zap.String("username", req.Username))
		return errors.New("用户名长度需在3-50位之间")
	}
	if len(req.Password) < 6 || len(req.Password) > 20 {
		logger.Warn("注册失败：密码长度不符合要求", zap.String("username", req.Username))
		return errors.New("密码长度需在6-20位之间")
	}

	// 2. 检查用户名是否已存在
	exist, err := s.userRepo.ExistByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("注册失败：查询用户名是否存在出错", zap.Error(err), zap.String("username", req.Username))
		return errors.New("注册失败：" + err.Error())
	}
	if exist {
		logger.Warn("注册失败：用户名已存在", zap.String("username", req.Username))
		return errors.New("用户名已存在")
	}
	exist, err = s.userRepo.ExistByUserEmail(ctx, req.Email)
	if err != nil {
		logger.Error("注册失败：查询邮箱是否存在出错", zap.Error(err), zap.String("email", req.Email))
		return errors.New("注册失败：" + err.Error())
	}
	if exist {
		logger.Warn("注册失败：邮箱已存在", zap.String("email", req.Email))
		return errors.New("邮箱已存在" + err.Error())
	}

	// 3. 构建用户实体（密码会被DAO层BeforeSave钩子加密）
	user := &model.User{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Role:     "user", // 默认普通用户
		Status:   1,      // 默认启用
	}

	// 4. 调用DAO创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("注册失败：创建用户出错", zap.Error(err), zap.String("username", req.Username))
		return errors.New("注册失败：" + err.Error())
	}

	logger.Info("用户注册成功", zap.String("username", req.Username))
	return nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *model.UserLoginRequest) (string, error) {
	// 1. 业务规则校验：参数非空
	if req.Username == "" || req.Password == "" {
		logger.Warn("登录失败：用户名/密码为空", zap.String("username", req.Username))
		return "", errors.New("用户名和密码不能为空")
	}

	// 2. 查询用户（包含密码）
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("登录失败：查询用户出错", zap.Error(err), zap.String("username", req.Username))
		return "", errors.New("用户名或密码错误")
	}
	if user == nil {
		logger.Warn("登录失败：用户不存在", zap.String("username", req.Username))
		return "", errors.New("用户名或密码错误")
	}

	// 3. 检查用户状态（禁用则拒绝登录）
	if user.Status != 1 {
		logger.Warn("登录失败：用户已被禁用", zap.String("username", req.Username))
		return "", errors.New("用户已被禁用，请联系管理员")
	}

	// 4. 验证密码（model.User的CheckPassword方法）
	if !user.CheckPassword(req.Password) {
		logger.Warn("登录失败：密码错误", zap.String("username", req.Username))
		return "", errors.New("用户名或密码错误")
	}

	// 5. 生成JWT Token（user.ID是int64，转int传给JWT）
	token, err := s.jwtService.GenerateToken(int(user.ID), user.Username, user.Role)
	if err != nil {
		logger.Error("登录失败：生成Token出错", zap.Error(err), zap.Int64("user_id", user.ID))
		return "", errors.New("登录失败：" + err.Error())
	}

	logger.Info("用户登录成功", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
	return token, nil
}

// Logout 用户登出（Token加入Redis黑名单）
func (s *userService) Logout(ctx context.Context, token string) error {
	// 1. 校验Token非空
	if token == "" {
		logger.Warn("登出失败：Token为空")
		return errors.New("Token不能为空")
	}

	// 2. 调用JWTService将Token加入黑名单
	if err := s.jwtService.BlacklistToken(token); err != nil {
		logger.Error("登出失败：Token加入黑名单出错", zap.Error(err), zap.String("token", token))
		return errors.New("登出失败：" + err.Error())
	}

	logger.Info("用户登出成功", zap.String("token", token))
	return nil
}
