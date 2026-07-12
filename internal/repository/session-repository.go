package repository

import (
	"time"

	"go-todo/internal/model"

	"gorm.io/gorm"
)

// SessionRepository 是一个结构体，封装了对用户会话的数据库操作
type SessionRepository struct {
	db *gorm.DB
}

// 创建 SessionRepository 实例。
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// 创建一条登录会话。
func (r *SessionRepository) Create(session *model.UserSession) error {
	return r.db.Create(session).Error
}

// 根据 session ID 查询登录会话。
func (r *SessionRepository) GetByID(id string) (*model.UserSession, error) {
	var session model.UserSession

	if err := r.db.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// 更新当前有效的 refresh token hash，并记录使用时间。
func (r *SessionRepository) RotateRefreshToken(
	id string,
	currentRefreshTokenHash string, // 当前的 refresh token hash
	newRefreshTokenHash string, // 新的 refresh token hash
	lastUsedAt time.Time,
) (bool, error) {
	result := r.db.
		Model(&model.UserSession{}).
		Where("id = ?", id).                                      // 更新指定的会话
		Where("refresh_token_hash = ?", currentRefreshTokenHash). // 确保当前的 refresh token hash 匹配，不被并行干扰
		Where("revoked_at IS NULL").
		Where("expires_at > ?", lastUsedAt).
		Updates(map[string]any{
			"refresh_token_hash": newRefreshTokenHash,
			"last_used_at":       lastUsedAt,
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

// 撤销一条登录会话。
func (r *SessionRepository) Revoke(
	id string,
	revokedAt time.Time,
) (bool, error) {
	result := r.db.
		Model(&model.UserSession{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}
