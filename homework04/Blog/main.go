package main

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"BlogSystem/config"
	"BlogSystem/handlers"
	"BlogSystem/middleware"
	"BlogSystem/models"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db, err := config.OpenDB("blog.db")
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
	config.SetDBPool(db)

	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		log.Fatal("数据库迁移失败: ", err)
	}

	r := gin.Default()
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())
	r.Use(middleware.Gzip())
	r.Use(middleware.StaticCache())

	r.Static("/static", "./static")

	templateDir := config.GetTemplateDir()
	r.LoadHTMLGlob(filepath.Join(templateDir, "*.html"))

	handlers.RegisterRoutes(r, db)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("服务启动: http://localhost:8080/register")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
