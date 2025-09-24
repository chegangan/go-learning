package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql" // 导入MySQL驱动
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ===== 全局变量和结构体 =====

var db *sql.DB
var jwtKey = []byte("my_secret_key_change_it_in_production") // 【重要】在生产环境中应使用更复杂的密钥

// User 结构体，增加了 Password 字段用于接收注册请求
type User struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	City     string `json:"city"`
	Password string `json:"password,omitempty"` // omitempty: 在JSON编码时如果为空则忽略此字段
}

// Product 结构体
type Product struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

// Claims JWT 的载荷结构体
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ===== 数据库相关函数 =====

func initDB() {
	dsn := "root:1234@tcp(127.0.0.1:3306)/go_http_learning?parseTime=true"
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("无法打开数据库: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	fmt.Println("数据库连接成功!")
}

// dbRegisterUser 在数据库中创建新用户，并哈希密码
func dbRegisterUser(u User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (name, age, city, password_hash) VALUES (?, ?, ?, ?)", u.Name, u.Age, u.City, string(hashedPassword))
	return err
}

// dbGetUserForLogin 从数据库获取用户信息用于登录验证
func dbGetUserForLogin(name string) (*User, string, error) {
	var u User
	var passwordHash string
	row := db.QueryRow("SELECT name, age, city, password_hash FROM users WHERE name = ?", name)
	if err := row.Scan(&u.Name, &u.Age, &u.City, &passwordHash); err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

// dbGetAllUsers 从数据库获取所有用户
func dbGetAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT name, age, city FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Name, &u.Age, &u.City); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// dbGetUserByName 从数据库按姓名获取单个用户
func dbGetUserByName(name string) (*User, error) {
	var u User
	row := db.QueryRow("SELECT name, age, city FROM users WHERE name = ?", name)
	if err := row.Scan(&u.Name, &u.Age, &u.City); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到用户，不是一个错误
		}
		return nil, err
	}
	return &u, nil
}

// dbGetAllProducts 从数据库获取所有商品
func dbGetAllProducts() ([]Product, error) {
	rows, err := db.Query("SELECT id, name, price, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// ===== 认证相关处理器 (Auth Handlers) =====

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}
	if u.Password == "" || u.Name == "" {
		http.Error(w, "用户名和密码不能为空", http.StatusBadRequest)
		return
	}

	err := dbRegisterUser(u)
	if err != nil {
		// 检查是否是重复键错误
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			http.Error(w, "用户名已存在", http.StatusConflict)
			return
		}
		http.Error(w, "创建用户失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "用户 %s 注册成功", u.Name)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	user, passwordHash, err := dbGetUserForLogin(creds.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
			return
		}
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password)); err != nil {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}

	// 密码正确，生成JWT
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "创建Token失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

// ===== 业务 API 处理器 =====

// usersRootHandler 处理 /api/users 的 GET 请求
func usersRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		users, err := dbGetAllUsers()
		if err != nil {
			http.Error(w, "获取用户失败", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(users)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// userSpecificHandler 处理 /api/users/{name} 的 GET 请求
func userSpecificHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if name == "" {
		http.Error(w, "缺少用户名", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := dbGetUserByName(name)
		if err != nil {
			http.Error(w, "查询用户失败", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(user)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// productsRootHandler 处理 /api/products 的 GET 请求
func productsRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		products, err := dbGetAllProducts()
		if err != nil {
			http.Error(w, "获取商品失败", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(products)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

// ===== 中间件 (Middleware) =====

// 【新增】loggingMiddleware 记录每个请求的日志
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("请求进入: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// 【新增】timingMiddleware 记录每个请求的处理时间
// 中间件接收一个 http.Handler，返回一个新的 http.Handler
// 在新的处理器中，先记录开始时间，调用下一个处理器，最后计算并记录耗时，从而达到包装的效果
func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("请求处理完毕: %s %s, 耗时: %v", r.Method, r.URL.Path, duration)
	})
}

// jwtAuthMiddleware 验证JWT的中间件
func jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "缺少认证头", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "无效的签名", http.StatusUnauthorized)
				return
			}
			http.Error(w, "无效的Token", http.StatusBadRequest)
			return
		}

		if !token.Valid {
			http.Error(w, "无效的Token", http.StatusUnauthorized)
			return
		}

		// Token有效，继续处理请求
		next.ServeHTTP(w, r)
	})
}

// ===== 主函数 (main) =====

func main() {
	initDB()
	defer db.Close()

	// http下还自带了一个默认的多路复用器DefaultServeMux，http.Handle和http.HandleFunc会注册路由到这个默认的多路复用器
	// --- 路由模块化 ---
	authMux := http.NewServeMux() // 公开的，用于认证
	apiMux := http.NewServeMux()  // 受保护的，用于业务API

	// 1. 注册认证路由到 authMux (无需JWT)
	authMux.HandleFunc("/auth/register", registerHandler)
	authMux.HandleFunc("/auth/login", loginHandler)

	// 2. 注册业务API路由到 apiMux (需要JWT)
	apiMux.HandleFunc("/api/users", usersRootHandler)
	apiMux.HandleFunc("/api/users/", userSpecificHandler)
	apiMux.HandleFunc("/api/products", productsRootHandler)

	// 3. 创建主路由器
	// 路由器也是一个handle，他的功能是根据请求路径选择其他的handle来处理请求
	// 在处理apimux这个handle之前，先调用jwtAuthMiddleware这个中间件，包装了apimux方法
	// jwtAuthMiddleware返回的也是一个handle，他是一个新的handle，这个handle会先验证jwt，然后再调用apimux
	mainMux := http.NewServeMux()
	mainMux.Handle("/auth/", authMux)
	// 将整个 /api/ 路径下的路由都用JWT中间件保护起来
	mainMux.Handle("/api/", jwtAuthMiddleware(apiMux))

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	mainMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 应用全局中间件（日志、计时）
	chainedHandler := loggingMiddleware(timingMiddleware(mainMux))

	// 自定义服务器
	// http下面自带了一个默认的server，ListenAndServe会启动这个默认的server
	server := &http.Server{
		Addr:         ":8000",
		Handler:      chainedHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	fmt.Println("服务器即将在 http://localhost:8000 启动")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
