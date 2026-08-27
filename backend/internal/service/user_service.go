package service

import (
	"errors"
	"log"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/utils"
)

var (
	ErrInvalidPhone       = errors.New("invalid phone number")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserDisabled       = errors.New("user disabled")
)

type UserService struct {
	userRepo   *repository.UserRepository
	configRepo *repository.SystemConfigRepository
}

func NewUserService(userRepo *repository.UserRepository, configRepo *repository.SystemConfigRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		configRepo: configRepo,
	}
}

// Register 用户注册
func (s *UserService) Register(phone, password string) (*model.PlatformUser, error) {
	// 检查手机号是否已存在
	exists, err := s.userRepo.CheckPhoneExists(phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// API Key 注册时自动生成（唯一），API Secret 需完成账户实名后再生成下发

	// 获取系统KYC价格（默认1.00）
	kycPrice := 1.00
	if s.configRepo != nil {
		priceStr, err := s.configRepo.GetConfig("kyc_price")
		if err == nil && priceStr != "" {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				kycPrice = price
			}
		}
	}

	log.Printf("新用户注册，设置KYC单价: %.2f元", kycPrice)

	// 创建用户（API Key 注册时自动生成；API Secret 实名成功后生成）
	user := &model.PlatformUser{
		Phone:         phone,
		PasswordHash:  string(hashedPassword),
		Balance:       0,
		APIKey:        utils.GenerateRandomKey(32),
		APISecret:     "",
		IsKYCVerified: 0,
		KYCPrice:      kycPrice, // 设置KYC单价（下游 API 业务调用扣费）
		Status:        1,
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(phone, password string) (*model.PlatformUser, error) {
	// 查询用户
	user, err := s.userRepo.GetUserByPhone(phone)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}

	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLoginTime(user.ID)

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int64) (*model.PlatformUser, error) {
	return s.userRepo.GetUserByID(id)
}

// GetUserByPhone 根据手机号获取用户
func (s *UserService) GetUserByPhone(phone string) (*model.PlatformUser, error) {
	return s.userRepo.GetUserByPhone(phone)
}

// GetByAPIKey 根据API Key获取用户
func (s *UserService) GetByAPIKey(apiKey string) (*model.PlatformUser, error) {
	return s.userRepo.GetByAPIKey(apiKey)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID int64, newPassword string) error {
	// 密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateUserPassword(userID, string(passwordHash))
}

// ResetAPIKey 重置API密钥（需先完成账户实名认证，Key 与 Secret 一并重新生成）
func (s *UserService) ResetAPIKey(userID int64) (string, string, error) {
	apiKey := utils.GenerateRandomKey(32)
	apiSecret := utils.GenerateRandomKey(32)

	err := s.userRepo.UpdateUserAPIKey(userID, apiKey, apiSecret)
	if err != nil {
		return "", "", err
	}

	return apiKey, apiSecret, nil
}

// UpdateKYCInfo 更新用户实名信息
func (s *UserService) UpdateKYCInfo(userID int64, name, idCard string) error {
	// 注：身份证号的加密在 AuthService.CreateAuth 中已实现
	// 此方法用于更新已认证用户的信息，建议生产环境禁用或添加严格权限控制
	return s.userRepo.UpdateUserKYCInfo(userID, name, idCard)
}
