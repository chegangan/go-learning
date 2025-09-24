package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"go-learning/learn-gohttp/version4/internal/handler" // 导入 handler
	"go-learning/learn-gohttp/version4/internal/store"   // 导入 store

	_ "github.com/go-sql-driver/mysql"
)

// 初始化依赖（数据库连接）。

// 创建各个层的实例 (store, handler)。

// 将依赖注入到需要它们的地方。

// 设置路由。

// 启动服务器。
func main() {
	// 1. 初始化依赖 (数据库)
	dsn := "root:your_password@tcp(127.0.0.1:3306)/webapp_db?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("无法打开数据库: %v", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	fmt.Println("数据库连接成功!")

	// 2. 创建 store 和 handler 实例，并注入依赖
	userStore := &store.UserStore{DB: db}
	productStore := &store.ProductStore{DB: db}

	authHandler := &handler.AuthHandler{UserStore: userStore}
	userHandler := &handler.UserHandler{UserStore: userStore}
	productHandler := &handler.ProductHandler{ProductStore: productStore}

	// 3. 设置路由
	authMux := http.NewServeMux()
	authMux.HandleFunc("/auth/register", authHandler.RegisterHandler)
	authMux.HandleFunc("/auth/login", authHandler.LoginHandler)

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/users", userHandler.RootHandler)
	apiMux.HandleFunc("/api/users/", userHandler.SpecificHandler)
	apiMux.HandleFunc("/api/products", productHandler.RootHandler)

	mainMux := http.NewServeMux()
	mainMux.Handle("/auth/", authMux)
	mainMux.Handle("/api/", handler.JwtAuthMiddleware(apiMux)) // 中间件在 handler 包里

	// 4. 应用全局中间件并配置服务器
	chainedHandler := handler.LoggingMiddleware(handler.TimingMiddleware(mainMux))
	server := &http.Server{
		Addr:    ":8000",
		Handler: chainedHandler,
		// ... 其他服务器配置 ...
	}

	// 5. 启动服务器
	fmt.Println("服务器即将在 http://localhost:8000 启动")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
