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
	Completed *bool
	Limit     int
	Offset    int
}

// Create 创建一个新的 Todo 任务
func (r *TodoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}

// GetByID 根据 ID 获取 Todo 任务
func (r *TodoRepository) GetByID(id int64) (*model.Todo, error) {
	var todo model.Todo

	if err := r.db.First(&todo, id).Error; err != nil {
		return nil, err
	}

	return &todo, nil
}

// applyFilters是总数和列表查询的公共方法，用于根据传入的参数应用过滤条件
func (r *TodoRepository) applyFilters(query *gorm.DB, params ListTodosParams) *gorm.DB {
	// 如果 Completed 参数不为 nil，则应用过滤条件
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
func (r *TodoRepository) Update(id int64, title string, description string) (int64, error) {
	result := r.db.
		Model(&model.Todo{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"title":       title,
			"description": description,
		})

	return result.RowsAffected, result.Error
}

// 标记完成
func (r *TodoRepository) MarkCompleted(id int64) error {
	return r.db.
		Model(&model.Todo{}).
		Where("id = ?", id).
		Update("completed", true).Error
}

// 删除任务
func (r *TodoRepository) Delete(id int64) error {
	return r.db.Delete(&model.Todo{}, id).Error
}
