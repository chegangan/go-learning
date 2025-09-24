package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync" // 引入 sync 包来处理并发访问
)

// -- 结构体定义 --
// 沿用之前的 User 结构体
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// -- 模拟数据库 --
// 为了演示，我们用一个内存中的切片来模拟数据库存储用户数据
// 使用 sync.RWMutex 来保证在并发读写 map 时的线程安全
var (
	users      = make(map[string]User) // 使用 map 来存储用户，用 Name 作为 key
	usersMutex = &sync.RWMutex{}       // 读写锁
)

// -- 处理器函数 --

// usersHandler 负责处理 /users 路径的请求
func usersHandler(w http.ResponseWriter, r *http.Request) {
	// 根据不同的 HTTP 方法 (GET, POST) 执行不同的逻辑
	switch r.Method {
	case http.MethodGet:
		getUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		// 如果是其他方法 (如 PUT, DELETE)，则返回 "方法不允许" 错误
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// getUsers 处理 GET 请求，返回所有用户列表
func getUsers(w http.ResponseWriter, r *http.Request) {
	usersMutex.RLock()         // 加读锁
	defer usersMutex.RUnlock() // 函数结束时解锁

	// 将 map 转换为切片，以便返回 JSON 数组
	var userList []User
	for _, user := range users {
		userList = append(userList, user)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(userList)
	if err != nil {
		log.Printf("编码 JSON 失败: %v", err)
	}
}

// createUser 处理 POST 请求，创建一个新用户
func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User

	// 【新知识点 1: 解析请求体中的 JSON】
	// json.NewDecoder 从请求体 (r.Body) 中读取数据
	// .Decode(&newUser) 将读取到的 JSON 数据解码并填充到 newUser 变量中
	// 这里的decode的参数是一个指向 User 结构体的指针，因为解码器需要修改这个变量的值，需要传递地址
	// 否则在函数内修改的是副本，在函数执行完之后副本就会销毁，外部的user变量不会被修改
	// 前面的encode的参数是user结构体的值，因为编码器只需要读取这个变量的值，golang是值传递，传递的是副本
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		// 如果 JSON 格式错误，返回一个 400 Bad Request 错误
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	// 简单的验证
	if newUser.Name == "" {
		http.Error(w, "姓名不能为空", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()         // 加写锁
	defer usersMutex.Unlock() // 函数结束时解锁

	// 检查用户是否已存在
	if _, exists := users[newUser.Name]; exists {
		http.Error(w, "用户已存在", http.StatusConflict) // 409 Conflict
		return
	}

	// 将新用户存入我们的“数据库”
	users[newUser.Name] = newUser
	log.Printf("新用户已添加: %s", newUser.Name)

	// 返回一个 201 Created 状态码，表示资源创建成功
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "用户 %s 创建成功!", newUser.Name)
}

// -- 主函数 --
func main() {
	// 【新知识点 2: 创建和使用 http.ServeMux】
	// 1. 不再使用全局的 http.HandleFunc，而是创建一个新的 ServeMux 实例
	mux := http.NewServeMux()

	// 2. 将路由处理器注册到我们自己的 mux 实例上
	mux.HandleFunc("/users", usersHandler)

	// 同样，我们也可以把之前的静态文件服务也注册到这个 mux 上
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("服务器即将在 http://localhost:8000 启动")
	// 3. 将我们创建的 mux 作为第二个参数传入 ListenAndServe
	// 这样，服务器就会使用我们定义的路由规则，而不是全局默认的
	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
