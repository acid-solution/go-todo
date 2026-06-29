package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 任务业务模型，包含后端自己处理业务需要的信息，可以用于前端，也可以用于数据库
type Todo struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;type:bigint unsigned;column:id"`
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

// 查询请求体
type ListTodoRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Completed string `form:"completed"`
}

// 统一响应体
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// 任务响应模型，就是data里应该包着返回给前端的模型，通常是由业务模型转换来的
type TodoResponse struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 列表响应体
type TodoListResponse struct {
	Items      []TodoResponse `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
}

// 数据库连接对象，只是一个入口
var db *sql.DB
var gormDB *gorm.DB

// 初始化数据库连接
func initDB() (*sql.DB, *gorm.DB) {
	dsn := "root:root123@tcp(127.0.0.1:3306)/go_todo?charset=utf8mb4&parseTime=True&loc=Local"

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接 MySQL 失败:", err)
	}

	if err := gdb.AutoMigrate(&Todo{}); err != nil {
		log.Fatal("AutoMigrate 失败:", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatal("获取底层 sql.DB 失败:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Ping MySQL 失败:", err)
	}

	return sqlDB, gdb
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
	var req ListTodoRequest
	//从前端请求中绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, http.StatusBadRequest, "查询参数无效")
		return
	}
	//如果没有传入分页参数，设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	//校验参数是否合法
	if req.Page < 1 {
		fail(c, http.StatusBadRequest, "page 必须大于等于 1")
		return
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		fail(c, http.StatusBadRequest, "page_size 必须在 1 到 100 之间")
		return
	}
	if req.Completed != "" && req.Completed != "true" && req.Completed != "false" {
		fail(c, http.StatusBadRequest, "completed 只能是 true 或 false")
		return
	}

	//查询总数
	where := ""
	whereArgs := make([]any, 0)

	if req.Completed != "" {
		completed := req.Completed == "true"
		where = " WHERE completed = ?"
		whereArgs = append(whereArgs, completed)
	}
	// 拼接查询总数的 SQL 语句
	countQuery := "SELECT COUNT(*) FROM todos" + where
	var total int64
	// 执行查询总数的 SQL 语句，并将结果扫描到 total 变量中
	if err := db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
		fail(c, http.StatusInternalServerError, "查询任务总数失败")
		return
	}
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	//计算偏移量
	offset := (req.Page - 1) * req.PageSize
	//拼接上半
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
	`
	// 创建一个切片来存储查询参数
	args := make([]any, 0)
	// 如果传入了 completed 参数，则添加 WHERE 子句和查询参数
	if req.Completed != "" {
		completed := req.Completed == "true"
		query += `
			WHERE completed = ?
		`
		args = append(args, completed)
	}
	// 拼接下半
	query += `
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`
	// 收集所有参数
	args = append(args, req.PageSize, offset)
	// 使用参数和拼接好的SQL语句查询数据库，需要用args...来解包切片
	rows, err := db.Query(query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询任务列表失败")
		return
	}
	//rows是一个迭代器，是从数据库一条一条读取数据的，所以要记得关闭
	defer rows.Close()

	//创建一个业务模型切片来存储任务列表
	todos := make([]*Todo, 0)
	//用rows.Next()迭代器来遍历查询结结果集
	for rows.Next() {
		var todo Todo
		//用rows.Scan()来把查询结果集的每一行数据扫描到业务模型中
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		//如果扫描失败，返回错误信息
		if err != nil {
			fail(c, http.StatusInternalServerError, "解析任务列表失败")
			return
		}
		//把业务模型添加到切片中
		todos = append(todos, &todo)
	}
	//检查迭代器是否有报错，迭代器报错是有没有正确完成迭代，任务错误在内部的scan就处理了
	if err := rows.Err(); err != nil {
		fail(c, http.StatusInternalServerError, "读取任务列表失败")
		return
	}

	//创建一个响应模型切片来返回给前端
	items := make([]TodoResponse, 0, len(todos))
	for _, todo := range todos {
		items = append(items, toTodoResponse(todo))
	}
	//手动封装响应体，返回给前端
	success(c, TodoListResponse{
		Items:      items,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

// 更新任务函数
func updateTodo(c *gin.Context) {
	var idReq TodoIDRequest

	// 获取路径参数 id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}

	var req UpdateTodoRequest

	// 获取任务名称和任务描述
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "任务名称不能为空")
		return
	}

	// 更新数据库
	result, err := db.Exec(
		`UPDATE todos
		 SET title = ?, description = ?
		 WHERE id = ?`,
		req.Title,
		req.Description,
		idReq.ID,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, "更新任务失败")
		return
	}

	// 表示刚才的更新更新了多少行
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fail(c, http.StatusInternalServerError, "获取更新结果失败")
		return
	}
	// 如果没有更新任何行，说明任务不存在，返回404
	if rowsAffected == 0 {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 查询更新后的最新任务
	todo, err := getTodoByID(idReq.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询更新后的任务失败")
		return
	}

	success(c, toTodoResponse(todo))
}

// 标记完成函数
func completeTodo(c *gin.Context) {
	var idReq TodoIDRequest

	// 获取路径参数 id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}

	// 先查询任务是否存在
	_, err := getTodoByID(idReq.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fail(c, http.StatusNotFound, "任务不存在")
			return
		}

		fail(c, http.StatusInternalServerError, "查询任务失败")
		return
	}

	// 更新数据库，将任务标记为已完成
	_, err = db.Exec(
		`UPDATE todos
		 SET completed = 1
		 WHERE id = ?`,
		idReq.ID,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, "标记任务完成失败")
		return
	}

	// 查询更新后的最新任务
	todo, err := getTodoByID(idReq.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询更新后的任务失败")
		return
	}

	success(c, toTodoResponse(todo))
}

// 删除任务函数
func deleteTodo(c *gin.Context) {
	var idReq TodoIDRequest

	// 获取路径参数 id
	if err := c.ShouldBindUri(&idReq); err != nil {
		fail(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}

	// 删除数据库中的任务
	result, err := db.Exec(
		`DELETE FROM todos
		 WHERE id = ?`,
		idReq.ID,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, "删除任务失败")
		return
	}

	// 判断是否真的删除了数据
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fail(c, http.StatusInternalServerError, "获取删除结果失败")
		return
	}

	if rowsAffected == 0 {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	success(c, nil)
}

func main() {
	db, gormDB = initDB()
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
