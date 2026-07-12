package service

import (
	"errors"
	"go-todo/internal/model"
	"go-todo/internal/repository"

	"gorm.io/gorm"
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
func (s *TodoService) CreateTodo(
	userID uint64,
	input CreateTodoInput,
) (*model.Todo, error) {
	todo := model.Todo{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
	}

	if err := s.repo.Create(&todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

// 根据ID获取任务
func (s *TodoService) GetTodoByID(
	userID uint64,
	id int64,
) (*model.Todo, error) {
	todo, err := s.repo.GetByID(userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}

		return nil, err
	}

	return todo, nil
}

// 根据条件查询任务列表
func (s *TodoService) ListTodos(
	userID uint64,
	input ListTodosInput,
) (*ListTodosResult, error) {
	var completed *bool

	if input.Completed != "" {
		value := input.Completed == "true"
		completed = &value
	}

	offset := (input.Page - 1) * input.PageSize

	params := repository.ListTodosParams{
		UserID:    userID,
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

	totalPages := int(
		(total + int64(input.PageSize) - 1) /
			int64(input.PageSize),
	)

	return &ListTodosResult{
		Items:      todos,
		Page:       input.Page,
		PageSize:   input.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// 更新任务
func (s *TodoService) UpdateTodo(
	userID uint64,
	id int64,
	input UpdateTodoInput,
) (*model.Todo, error) {
	if _, err := s.GetTodoByID(userID, id); err != nil {
		return nil, err
	}

	if err := s.repo.Update(
		userID,
		id,
		input.Title,
		input.Description,
	); err != nil {
		return nil, err
	}

	return s.GetTodoByID(userID, id)
}

// 标记完成
func (s *TodoService) CompleteTodo(
	userID uint64,
	id int64,
) (*model.Todo, error) {
	if _, err := s.GetTodoByID(userID, id); err != nil {
		return nil, err
	}

	if err := s.repo.MarkCompleted(userID, id); err != nil {
		return nil, err
	}

	return s.GetTodoByID(userID, id)
}

// 删除任务
func (s *TodoService) DeleteTodo(
	userID uint64,
	id int64,
) error {
	if _, err := s.GetTodoByID(userID, id); err != nil {
		return err
	}

	return s.repo.Delete(userID, id)
}
