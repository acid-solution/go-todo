package repository

import (
	"go-todo/internal/model"

	"gorm.io/gorm"
)

// UserRepository 封装 users 表的数据库操作。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create 创建新用户。
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByUsername 根据用户名查询用户。
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User

	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
