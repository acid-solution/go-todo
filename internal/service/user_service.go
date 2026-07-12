package service

import (
	"errors"
	"go-todo/internal/model"
	"go-todo/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo        *repository.UserRepository
	sessionRepo     *repository.SessionRepository
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewUserService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidSession        = errors.New("invalid session")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

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

// 登录成功返回结果
type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	SessionID    string
}

// 刷新需要什么
type RefreshInput struct {
	SessionID    string
	RefreshToken string
}

// 刷新成功返回结果
type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
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
	if err := s.userRepo.Create(&user); err != nil {
		if errors.Is(err, repository.ErrUsernameAlreadyExists) {
			return nil, ErrUsernameAlreadyExists
		}

		return nil, err
	}

	return &user, nil
}

// 登录用户
func (s *UserService) Login(input LoginInput) (*LoginResult, error) {
	// 1. 根据用户名获取用户
	user, err := s.userRepo.GetByUsername(input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// 2. 校验密码
	if !CheckPassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// 3. 生成 sessionID
	sessionID := uuid.NewString()

	// 4. 生成 refresh token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 5. 哈希 refresh token
	refreshTokenHash, err := HashPassword(refreshToken)
	if err != nil {
		return nil, err
	}

	// 6. 创建 session
	now := time.Now()
	session := model.UserSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        now.Add(s.refreshTokenTTL),
	}

	if err := s.sessionRepo.Create(&session); err != nil {
		return nil, err
	}

	// 7. 生成 access token
	accessToken, err := GenerateAccessToken(user.ID, sessionID, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	// 8. 返回完整登录结果
	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

// 刷新用户的 access token 和 refresh token
func (s *UserService) Refresh(input RefreshInput) (*RefreshResult, error) {
	// 1. 根据 session ID 查询登录会话
	session, err := s.sessionRepo.GetByID(input.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidSession
		}

		return nil, err
	}

	now := time.Now()

	// 2. 检查 session 是否已经被撤销
	if session.RevokedAt != nil {
		return nil, ErrInvalidSession
	}

	// 3. 检查 refresh session 是否已经过期
	if !now.Before(session.ExpiresAt) {
		return nil, ErrInvalidSession
	}

	// 4. 校验 refresh token
	if !CheckPassword(input.RefreshToken, session.RefreshTokenHash) {
		return nil, ErrInvalidSession
	}

	// 5. 生成新的 refresh token
	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 6. 哈希新的 refresh token
	newRefreshTokenHash, err := HashPassword(newRefreshToken)
	if err != nil {
		return nil, err
	}

	// 7. 生成新的 access token
	newAccessToken, err := GenerateAccessToken(
		session.UserID,
		session.ID,
		s.jwtSecret,
		s.accessTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	// 8. 轮换新的 refresh token
	rotated, err := s.sessionRepo.RotateRefreshToken(
		session.ID,
		session.RefreshTokenHash,
		newRefreshTokenHash,
		now,
	)
	if err != nil {
		return nil, err
	}

	if !rotated {
		return nil, ErrInvalidSession
	}

	// 9. 返回新的 token
	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		SessionID:    session.ID,
	}, nil
}

func (s *UserService) Logout(sessionID string) error {
	now := time.Now()

	revoked, err := s.sessionRepo.Revoke(sessionID, now)
	if err != nil {
		return err
	}

	if !revoked {
		return ErrInvalidSession
	}

	return nil
}
