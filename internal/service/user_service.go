package service

import (
	"errors"
	"go-todo/internal/model"
	"go-todo/internal/repository"

	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// 注册需要什么
type RegisterInput struct {
	Username string
	Password string
}

// 登录需要什么
type LoginInput struct {
	Username string
	Password string
}

// 注册用户
func (s *UserService) Register(input RegisterInput) (*model.User, error) {
	// 1. 生成密码哈希
	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	// 2. 封装用户数据
	user := model.User{
		Username:     input.Username,
		PasswordHash: passwordHash,
	}
	// 3. 保存用户到数据库
	if err := s.repo.Create(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// 登录用户
func (s *UserService) Login(input LoginInput) (*model.User, error) {
	// 1. 根据用户名获取用户
	user, err := s.repo.GetByUsername(input.Username)
	// 2. 检查用户名是否存在
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}
	// 3. 检查密码是否正确
	if !CheckPassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
