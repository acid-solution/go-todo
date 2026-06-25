package main

import (
	"database/sql"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 业务模型，包含后端自己处理业务需要的信息，可以用于前端，也可以用于数据库
type Todo struct {
	ID          int64
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// 响应模型，就是data里应该包着返回给前端的模型，通常是由业务模型转换来的
type TodoResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 创建请求体
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// 更新请求体
type UpdateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// 路径参数请求体
type TodoIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

// 统一响应体
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// 用内存模拟数据库
type TodoStore struct {
	todos  map[int64]*Todo
	nextID int64
}

// 构造函数
func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos:  make(map[int64]*Todo),
		nextID: 1,
	}
}

// 使用后端内存的临时测试用数据库
var store = NewTodoStore()

// 数据库连接对象，只是一个入口
var db *sql.DB

// 初始化数据库连接
func initDB() *sql.DB {
	//初始化dsn
	dsn := "root:root123@tcp(127.0.0.1:3306)/go_todo?charset=utf8mb4&parseTime=True&loc=Local"
	//连接数据库
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	//测试数据库连接是否成功
	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}

	return database
}

// 响应辅助函数
func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Code:    status,
		Message: message,
		Data:    nil,
	})
}

// 业务模型转换响应模型
func toTodoResponse(todo *Todo) TodoResponse {
	return TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

// 根据ID从数据库中获取任务
func getTodoByID(id int64) (*Todo, error) {
	var todo Todo

	err := db.QueryRow(
		`SELECT id, title, description, completed, created_at, updated_at
		 FROM todos
		 WHERE id = ?`,
		id,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// 创建接口函数
func createTodo(c *gin.Context) {
	var req CreateTodoRequest

	//从前端请求中绑定请求体数据并进行错误处理
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "任务名称不能为空")
		return
	}
	//把数据插入数据库
	result, err := db.Exec(
		"INSERT INTO todos (title, description) VALUES (?, ?)",
		req.Title,
		req.Description,
	)
	//如果插入失败，返回错误信息
	if err != nil {
		fail(c, http.StatusInternalServerError, "创建任务失败")
		return
	}
	//获取插入数据的ID
	id, err := result.LastInsertId()
	if err != nil {
		fail(c, http.StatusInternalServerError, "获取任务 ID 失败")
		return
	}
	//根据ID查询新建的任务
	todo, err := getTodoByID(id)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询新建任务失败")
		return
	}
	//返回响应模型
	success(c, toTodoResponse(todo))
}

// 列表返回函数
func listTodos(c *gin.Context) {
	//创建一个业务模型的列表（转json）用来装该返回的任务
	todos := make([]*Todo, 0, len(store.todos))

	//从数据库里取出数据装进列表
	for _, todo := range store.todos {
		todos = append(todos, todo)
	}

	//因为用map模拟数据库，所以需要排序一下
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].ID < todos[j].ID
	})

	//把业务模型转换成响应模型
	resp := make([]TodoResponse, 0, len(todos))
	for _, todo := range todos {
		resp = append(resp, toTodoResponse(todo))
	}

	success(c, resp)
}

// 更新任务函数
func updateTodo(c *gin.Context) {
	var idReq TodoIDRequest
	//获取id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}

	var req UpdateTodoRequest
	//获取任务名称和任务描述
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "任务名称不能为空")
		return
	}
	//新建业务模型，查询数据库
	todo, exists := store.todos[idReq.ID]
	if !exists {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	//完善业务模型
	todo.Title = req.Title
	todo.Description = req.Description

	success(c, toTodoResponse(todo))
}

// 标记完成函数
func completeTodo(c *gin.Context) {
	var idReq TodoIDRequest
	//获取id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	//创建业务模型
	todo, exists := store.todos[idReq.ID]
	if !exists {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	//修改数据库
	todo.Completed = true

	success(c, toTodoResponse(todo))
}

// 删除任务函数
func deleteTodo(c *gin.Context) {
	var idReq TodoIDRequest
	//获取id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	//删除操作不需要业务模型，所以只查数据库就行
	if _, exists := store.todos[idReq.ID]; !exists {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	//操作数据库，这里是go里map的内置操作
	delete(store.todos, idReq.ID)

	success(c, nil)
}

func main() {
	db = initDB()
	defer db.Close()

	r := gin.Default()
	//用于向访问方开放文件
	r.Static("/static", "./static")
	//访问根目录的时候返回该文件
	r.StaticFile("/", "./static/index.html")

	r.POST("/api/todos", createTodo)
	r.GET("/api/todos", listTodos)
	//整体更新一般用put
	r.PUT("/api/todos/:id", updateTodo)
	//局部更新一般用patch
	r.PATCH("/api/todos/:id/done", completeTodo)
	r.DELETE("/api/todos/:id", deleteTodo)

	r.Run(":8081")
}
