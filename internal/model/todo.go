package model

import "time"

// 它被用作业务模型，但实际上已经混入了数据库模型的职责,毕竟要用aotomigrate创建表结构,所以这里的模型是数据库模型,也可以叫做数据库实体模型
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
