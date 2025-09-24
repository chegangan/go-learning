package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
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

	//把“做事”（checkURL）和“收集结果”这两个关注点分离开来了
	var wg sync.WaitGroup
	// 创建一个带缓冲的 channel，容量等于URL的数量
	resultsChannel := make(chan CheckResult, len(urls))

	fmt.Println("开始并发检查 (使用Channel)...")
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			// 将检查结果发送到 channel
			resultsChannel <- checkURL(u)
		}(url)
	}

	// 启动一个新的 goroutine，它专门负责等待所有检查完成后关闭 channel
	// 这样做可以防止主流程阻塞在 wg.Wait()，从而无法接收 channel 的数据
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	// 使用 for range 遍历 channel，直到 channel 被关闭
	// 这个循环会在这里阻塞，直到有数据可读或 channel 关闭
	for result := range resultsChannel {
		if result.Error != nil {
			fmt.Printf("URL: %s, 状态: 失败, 错误: %v\n", result.URL, result.Error)
		} else {
			fmt.Printf("URL: %s, 状态码: %d, 延迟: %v\n", result.URL, result.StatusCode, result.Latency)
		}
	}

	fmt.Println("检查完成！")
}
