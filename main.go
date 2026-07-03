package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

const (
	CodeOK              = 0
	CodeInvalidArgument = 10001
	CodeNotFound        = 10002
	CodeInternalError   = 20001
)

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

type Config struct {
	Port     string
	MySQLDSN string
}

func loadConfig() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:     getEnv("PORT", "8081"),
		MySQLDSN: getEnv("MYSQL_DSN", ""),
	}

	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN 未设置")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// 数据库连接对象，只是一个入口
var gormDB *gorm.DB
var logger *slog.Logger

// 初始化数据库连接
func initDB(dsn string) *gorm.DB {
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

	return gdb
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

const requestIDKey = "request_id"

// 生成request_id中间件，给每个请求生成一个唯一的request_id，并将其设置到请求上下文中，方便后续日志记录和追踪
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 request_id，如果没有则生成一个新的 UUID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		// 将 request_id 设置到请求上下文中，方便后续处理使用
		c.Set(requestIDKey, requestID)
		// 将 request_id 设置到响应头中，方便客户端获取
		c.Header("X-Request-ID", requestID)
		//纯前置中间件其实不需要调用c.Next()，但是为了保证中间件链的完整性，还是调用一下
		c.Next()
	}
}

// 日志记录中间件，记录每个请求的详细信息，包括请求方法、路径、状态码、耗时、客户端 IP 等
func requestLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		// 调用下一个中间件或处理函数，继续处理请求
		c.Next()
		// 其他中间件或处理函数执行完毕后
		// 记录请求结束时间
		latency := time.Since(start)
		// 获取响应状态码
		status := c.Writer.Status()
		// 从请求上下文中获取 request_id
		requestID := c.GetString(requestIDKey)
		//用键值对切片记录请求信息，方便后续日志记录和追踪
		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		}
		// 如果请求处理过程中有错误发生，则将错误信息添加到日志属性中
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}
		// 根据响应状态码的不同，记录不同级别的日志信息
		switch {
		case status >= http.StatusInternalServerError:
			// 500以上的状态码都算错误
			logger.Error("request completed", attrs...)
		case status >= http.StatusBadRequest:
			// 400-499的状态码都算警告
			logger.Warn("request completed", attrs...)
		default:
			// 200-399的状态码都算成功
			logger.Info("request completed", attrs...)
		}
	}
}

// 响应辅助函数
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeOK,
		Message: "success",
		Data:    data,
	})
}

func fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func failInvalidArgument(c *gin.Context, message string) {
	fail(c, http.StatusBadRequest, CodeInvalidArgument, message)
}

func failNotFound(c *gin.Context, message string) {
	fail(c, http.StatusNotFound, CodeNotFound, message)
}

func failInternalError(c *gin.Context, message string) {
	fail(c, http.StatusInternalServerError, CodeInternalError, message)
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
	//返回按主键查到的第一条数据，如果没有查到会返回ErrRecordNotFound错误
	if err := gormDB.First(&todo, id).Error; err != nil {
		return nil, err
	}

	return &todo, nil
}

// 创建接口函数
func createTodo(c *gin.Context) {
	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		failInvalidArgument(c, "任务名称不能为空")
		return
	}
	// 创建一个新的任务实例，并将请求体中的标题和描述赋值给它
	todo := Todo{
		Title:       req.Title,
		Description: req.Description,
	}
	// 使用 GORM 的 Create 方法将任务保存到数据库中，并检查是否有错误发生，GORM在创建后会自动填充ID和时间戳，不用再查一次了
	if err := gormDB.Create(&todo).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建任务失败")
		return
	}

	success(c, toTodoResponse(&todo))
}

// 列表返回函数
func listTodos(c *gin.Context) {
	var req ListTodoRequest
	// 绑定查询参数到请求体结构体
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, http.StatusBadRequest, "查询参数无效")
		return
	}
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}
	// 参数校验
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
	// 定义一个函数来决定是否需要应用过滤条件
	// 这个函数接受一个 GORM 查询对象，并根据请求体中的 completed 参数来应用过滤条件
	applyFilters := func(query *gorm.DB) *gorm.DB {
		if req.Completed != "" {
			completed := req.Completed == "true"
			query = query.Where("completed = ?", completed)
		}

		return query
	}
	// 查询任务总数
	var total int64
	if err := applyFilters(gormDB.Model(&Todo{})).Count(&total).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询任务总数失败")
		return
	}
	// 计算总页数和偏移量
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	offset := (req.Page - 1) * req.PageSize
	// 查询当前页任务列表
	var todos []Todo
	if err := applyFilters(gormDB.Model(&Todo{})).
		Order("created_at DESC").
		Order("id DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&todos).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询任务列表失败")
		return
	}
	// 将业务模型转换为响应模型
	items := make([]TodoResponse, 0, len(todos))
	for i := range todos {
		items = append(items, toTodoResponse(&todos[i]))
	}
	// 手动封装响应返回
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
		failInvalidArgument(c, "无效的任务 ID")
		return
	}

	var req UpdateTodoRequest
	// 绑定请求体 JSON 数据到 req 结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		failInvalidArgument(c, "任务名称不能为空")
		return
	}
	//Updates更新多个字段，传入一个map[string]interface{}，key是字段名，value是要更新的值
	result := gormDB.
		Model(&Todo{}).
		Where("id = ?", idReq.ID).
		Updates(map[string]any{
			"title":       req.Title,
			"description": req.Description,
		})

	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "更新任务失败")
		return
	}

	if result.RowsAffected == 0 {
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
		failInvalidArgument(c, "无效的任务 ID")
		return
	}
	// 查询任务是否存在
	_, err := getTodoByID(idReq.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "任务不存在")
			return
		}

		fail(c, http.StatusInternalServerError, "查询任务失败")
		return
	}
	// Update方法更新单个字段，传入字段名和要更新的值
	if err := gormDB.
		Model(&Todo{}).
		Where("id = ?", idReq.ID).
		Update("completed", true).Error; err != nil {
		fail(c, http.StatusInternalServerError, "标记任务完成失败")
		return
	}

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
		failInvalidArgument(c, "无效的任务 ID")
		return
	}
	// 查询任务是否存在
	_, err := getTodoByID(idReq.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "任务不存在")
			return
		}

		fail(c, http.StatusInternalServerError, "查询任务失败")
		return
	}
	// Delete方法按主键删除记录，传入要删除的模型和主键值
	if err := gormDB.Delete(&Todo{}, idReq.ID).Error; err != nil {
		fail(c, http.StatusInternalServerError, "删除任务失败")
		return
	}

	success(c, nil)
}

func main() {
	logger = initLogger()
	cfg := loadConfig()
	gormDB = initDB(cfg.MySQLDSN)

	r := gin.New()

	r.Use(requestIDMiddleware())
	r.Use(requestLoggerMiddleware(logger))
	r.Use(gin.Recovery())

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

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
