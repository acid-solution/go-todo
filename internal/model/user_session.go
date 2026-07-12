package model

import "time"

// UserSession 表示用户的一次登录会话。
type UserSession struct {
	ID               string     `gorm:"primaryKey;type:char(36);column:id"`
	UserID           uint64     `gorm:"type:bigint unsigned;not null;index;column:user_id"`
	RefreshTokenHash string     `gorm:"type:varchar(255);not null;column:refresh_token_hash"`
	ExpiresAt        time.Time  `gorm:"type:datetime;not null;column:expires_at"`
	LastUsedAt       *time.Time `gorm:"type:datetime;column:last_used_at"`
	RevokedAt        *time.Time `gorm:"type:datetime;column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
