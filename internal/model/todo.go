package model

import "time"

// 任务业务模型，包含后端自己处理业务需要的信息，可以用于前端，也可以用于数据库
type Todo struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned;column:id"`
	Title       string    `gorm:"type:varchar(255);not null;column:title"`
	Description string    `gorm:"type:varchar(1000);not null;default:'';column:description"`
	Completed   bool      `gorm:"not null;default:false;column:completed"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// GORM默认会把结构体名的复数形式作为表名，也可以手动指定表名为todos
func (Todo) TableName() string {
	return "todos"
}
