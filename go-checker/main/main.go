package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type CheckResult struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Error      error
}

func checkURL(url string) CheckResult {
	client := http.Client{
		Timeout: 5 * time.Second, // 设置一个5秒的超时，非常重要！
	}
	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{URL: url, Error: err}
	}
	defer resp.Body.Close()

	return CheckResult{URL: url, StatusCode: resp.StatusCode, Latency: latency}
}

func main() {
	// 1. 使用 flag 包接收命令行传入的文件名
	filePath := flag.String("file", "urls.txt", "包含URL列表的文件路径")
	flag.Parse()

	// 2. 读取并解析文件
	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urls []string
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("读取文件时发生错误: %v", err)
	}

	// 3. 顺序执行检查
	fmt.Println("开始检查...")
	for _, url := range urls {
		result := checkURL(url)
		if result.Error != nil {
			fmt.Printf("URL: %s, 状态: 失败, 错误: %v\n", result.URL, result.Error)
		} else {
			fmt.Printf("URL: %s, 状态码: %d, 延迟: %v\n", result.URL, result.StatusCode, result.Latency)
		}
	}
	fmt.Println("检查完成！")
}
