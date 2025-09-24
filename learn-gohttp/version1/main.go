package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// -- 结构体定义 --
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// -- 处理器函数 --

func rootHandler(w http.ResponseWriter, r *http.Request) {

	// Printf: Print formatted。将格式化字符串输出到标准输出（也就是你的控制台/终端）。
	// Sprintf: String print formatted。不直接打印，而是将格式化后的结果作为字符串返回。
	// Fprintf: File print formatted。将格式化字符串输出到指定的写入器 (Writer)。
	fmt.Fprintf(w, "你好，欢迎来到我的 Go Web 服务器！")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Hello, World!")
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "陌生人"
	}
	fmt.Fprintf(w, "你好, %s!", name)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		Name: "小明",
		Age:  25,
		City: "北京",
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Printf("编码 JSON 失败: %v", err)
	}

	// 另一种方式是先将对象转换为 JSON 字符串，然后再写入响应,这个方式常见于需要对json进行处理
	// 使用encode可以将对象用流的方式写入响应，对内存压力小。
	//     // 1. 使用 Marshal 将对象转换为 []byte
	// jsonData, err := json.Marshal(user)
	// if err != nil {
	//     http.Error(w, "无法编码 JSON", http.StatusInternalServerError)
	//     log.Printf("编码 JSON 失败: %v", err)
	//     return
	// }

	// // 2. 将 []byte 写入 ResponseWriter
	// w.Write(jsonData)
}

// -- 主函数 --

func main() {
	// 静态文件服务器
	// fs是一个文件服务器，处理/static/路径下的请求，提供static目录中的文件
	fs := http.FileServer(http.Dir("static"))
	// 将fs注册到/static/路径下，并去掉/static/前缀，所有对/static/的请求都会被路由到fs
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API 路由
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/greet", greetHandler)
	http.HandleFunc("/user", userHandler)

	// 启动服务器
	fmt.Println("服务器即将在 http://localhost:8000 启动")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
