package service

import (
	"errors"
	"go-todo/internal/model"
	"go-todo/internal/repository"
)

var ErrTodoNotFound = errors.New("todo not found")

// service依赖repository
type TodoService struct {
	repo *repository.TodoRepository
}

// NewTodoService 创建一个新的 TodoService 实例，包含 TodoRepository
func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// 定义输入和输出结构体
type CreateTodoInput struct {
	Title       string
	Description string
}

type UpdateTodoInput struct {
	Title       string
	Description string
}

type ListTodosInput struct {
	Page      int
	PageSize  int
	Completed string
}

type ListTodosResult struct {
	Items      []model.Todo
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

// 创建任务
func (s *TodoService) CreateTodo(input CreateTodoInput) (*model.Todo, error) {
	todo := model.Todo{
		Title:       input.Title,
		Description: input.Description,
	}

	if err := s.repo.Create(&todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

// 根据ID获取任务
func (s *TodoService) GetTodoByID(id int64) (*model.Todo, error) {
	return s.repo.GetByID(id)
}

// 根据条件查询任务列表
func (s *TodoService) ListTodos(input ListTodosInput) (*ListTodosResult, error) {
	var completed *bool

	if input.Completed != "" {
		value := input.Completed == "true"
		completed = &value
	}

	offset := (input.Page - 1) * input.PageSize

	params := repository.ListTodosParams{
		Completed: completed,
		Limit:     input.PageSize,
		Offset:    offset,
	}

	total, err := s.repo.Count(params)
	if err != nil {
		return nil, err
	}

	todos, err := s.repo.List(params)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(input.PageSize) - 1) / int64(input.PageSize))

	return &ListTodosResult{
		Items:      todos,
		Page:       input.Page,
		PageSize:   input.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// 更新任务
func (s *TodoService) UpdateTodo(id int64, input UpdateTodoInput) (*model.Todo, error) {
	rowsAffected, err := s.repo.Update(id, input.Title, input.Description)
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrTodoNotFound
	}

	return s.GetTodoByID(id)
}

// 标记完成
func (s *TodoService) CompleteTodo(id int64) (*model.Todo, error) {
	if _, err := s.GetTodoByID(id); err != nil {
		return nil, err
	}

	if err := s.repo.MarkCompleted(id); err != nil {
		return nil, err
	}

	return s.GetTodoByID(id)
}

// 删除任务
func (s *TodoService) DeleteTodo(id int64) error {
	if _, err := s.GetTodoByID(id); err != nil {
		return err
	}

	return s.repo.Delete(id)
}
