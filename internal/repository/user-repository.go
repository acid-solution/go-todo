package repository

import (
	"errors"
	"go-todo/internal/model"

	"github.com/go-sql-driver/mysql"
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

var ErrUsernameAlreadyExists = errors.New("username already exists")

// Create 创建新用户。
func (r *UserRepository) Create(user *model.User) error {
	err := r.db.Create(user).Error
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrUsernameAlreadyExists
	}

	return err
}

// GetByUsername 根据用户名查询用户。
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User

	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
