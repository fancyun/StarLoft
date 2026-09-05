package service

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/utils"
)

var (
	ErrInvalidPhone          = errors.New("invalid phone number")
	ErrPhoneAlreadyExists    = errors.New("phone already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrUserDisabled          = errors.New("user disabled")
)

// ValidateUsername 校验用户名：仅支持英文+数字+下划线，长度3-32
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}

// ValidateEmail 校验邮箱格式
func ValidateEmail(email string) bool {
	if len(email) < 5 || len(email) > 100 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

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

// Register 用户注册（手机号+用户名+邮箱必填；验证码已在 handler 层校验）
func (s *UserService) Register(phone, username, email, password string) (*model.PlatformUser, error) {
	// 校验用户名与邮箱格式
	if !ValidateUsername(username) {
		return nil, ErrInvalidUsername
	}
	if !ValidateEmail(email) {
		return nil, ErrInvalidEmail
	}

	// 检查手机号是否已存在
	exists, err := s.userRepo.CheckPhoneExists(phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.CheckUsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// API Key 注册时自动生成（唯一），API Secret 需完成账户实名后再生成下发

	// 创建用户（API Key 注册时自动生成；API Secret 实名成功后生成）
	user := &model.PlatformUser{
		Phone:         phone,
		Username:      username,
		Email:         email,
		PasswordHash:  string(hashedPassword),
		Balance:       0,
		APIKey:        utils.GenerateRandomKey(32),
		APISecret:     "",
		IsKYCVerified: 0,
		Status:        1,
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录（支持用户名/手机号/邮箱）
func (s *UserService) Login(account, password string) (*model.PlatformUser, error) {
	// 查询用户
	user, err := s.userRepo.GetUserByAccount(account)
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
