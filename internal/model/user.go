// internal/model/user.go
package model

import "time"

// User 表示系统用户。
// 注意：这里不存 Password，只存 PasswordHash。
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned;column:id"`
	Username     string    `gorm:"type:varchar(64);not null;uniqueIndex;column:username"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
