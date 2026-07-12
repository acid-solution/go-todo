package repository

import (
	"go-todo/internal/model"

	"gorm.io/gorm"
)

// TodoRepository 是一个结构体，封装了对 Todo 任务的数据库操作
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository 创建一个新的 TodoRepository 实例，包含数据库连接对象
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{
		db: db,
	}
}

// ListTodos 根据条件查询 Todo 任务列表
type ListTodosParams struct {
	UserID    uint64
	Completed *bool
	Limit     int
	Offset    int
}

// Create 创建一个新的 Todo 任务
func (r *TodoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}

// GetByID 根据 ID 获取 Todo 任务
func (r *TodoRepository) GetByID(
	userID uint64,
	id int64,
) (*model.Todo, error) {
	var todo model.Todo

	err := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&todo).
		Error
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// applyFilters是总数和列表查询的公共方法，用于根据传入的参数应用过滤条件
func (r *TodoRepository) applyFilters(
	query *gorm.DB,
	params ListTodosParams,
) *gorm.DB {
	query = query.Where("user_id = ?", params.UserID)

	if params.Completed != nil {
		query = query.Where("completed = ?", *params.Completed)
	}

	return query
}

// 统计任务总数
func (r *TodoRepository) Count(params ListTodosParams) (int64, error) {
	var total int64

	err := r.applyFilters(r.db.Model(&model.Todo{}), params).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

// 查询全部列表
func (r *TodoRepository) List(params ListTodosParams) ([]model.Todo, error) {
	var todos []model.Todo

	err := r.applyFilters(r.db.Model(&model.Todo{}), params).
		Order("created_at DESC").
		Order("id DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&todos).Error
	if err != nil {
		return nil, err
	}

	return todos, nil
}

// 根据id更新任务
func (r *TodoRepository) Update(
	userID uint64,
	id int64,
	title string,
	description string,
) error {
	return r.db.
		Model(&model.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"title":       title,
			"description": description,
		}).
		Error
}

// 标记完成
func (r *TodoRepository) MarkCompleted(
	userID uint64,
	id int64,
) error {
	return r.db.
		Model(&model.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("completed", true).
		Error
}

// 删除任务
func (r *TodoRepository) Delete(
	userID uint64,
	id int64,
) error {
	return r.db.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Todo{}).
		Error
}
