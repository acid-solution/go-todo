package database

import (
	"log"

	"go-todo/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMySQL(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接 MySQL 失败:", err)
	}

	if err := db.AutoMigrate(&model.Todo{}, &model.User{}); err != nil {
		log.Fatal("AutoMigrate 失败:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取底层 sql.DB 失败:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Ping MySQL 失败:", err)
	}

	return db
}
