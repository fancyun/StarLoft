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
	userRepo  *repository.UserRepository
	apiKeyRepo *repository.ApiKeyRepository
}

func NewUserService(userRepo *repository.UserRepository, apiKeyRepo *repository.ApiKeyRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

// Register 用户注册（手机号+用户名+邮箱必填；验证码已在 handler 层校验）
func (s *UserService) Register(phone, username, email, password string) (*model.User, error) {
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

	// 创建用户
	user := &model.User{
		Phone:         phone,
		Username:      username,
		Email:         email,
		PasswordHash:  string(hashedPassword),
		Balance:       0,
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
func (s *UserService) Login(account, password string) (*model.User, error) {
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
func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return s.userRepo.GetUserByID(id)
}

// GetUserByPhone 根据手机号获取用户
func (s *UserService) GetUserByPhone(phone string) (*model.User, error) {
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

// ResetAPIKey 重置用户 API 密钥（Key 与 Secret 一并重新生成，默认权限 all）
func (s *UserService) ResetAPIKey(userID int64) (string, string, error) {
	apiKey := utils.GenerateRandomKey(32)
	apiSecret := utils.GenerateRandomKey(32)

	if err := s.apiKeyRepo.DeleteByUser(userID); err != nil {
		return "", "", err
	}
	k := &model.ApiKey{
		UserID:     userID,
		APIKey:     apiKey,
		APISecret:  apiSecret,
		Permission: "all",
	}
	if err := s.apiKeyRepo.Create(k); err != nil {
		return "", "", err
	}

	return apiKey, apiSecret, nil
}

// EnsureAPIKey 实名为用户补发 API 密钥对（不存在时生成，默认权限 all）
func (s *UserService) EnsureAPIKey(userID int64) error {
	if _, err := s.apiKeyRepo.GetByUser(userID); err == nil {
		return nil
	}
	k := &model.ApiKey{
		UserID:     userID,
		APIKey:     utils.GenerateRandomKey(32),
		APISecret:  utils.GenerateRandomKey(32),
		Permission: "all",
	}
	return s.apiKeyRepo.Create(k)
}

// GetAPIKeyByUser 查询用户当前生效的 API 密钥
func (s *UserService) GetAPIKeyByUser(userID int64) (*model.ApiKey, error) {
	return s.apiKeyRepo.GetByUser(userID)
}

// SetAPIKeyPermission 设置用户 API 密钥权限范围（all 或单个服务标识）
func (s *UserService) SetAPIKeyPermission(userID int64, permission string) error {
	return s.apiKeyRepo.SetPermission(userID, permission)
}
