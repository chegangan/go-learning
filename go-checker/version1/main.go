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

	// var wg sync.WaitGroup // 声明一个 WaitGroup

	// fmt.Println("开始并发检查...")
	// for _, url := range urls {
	// 	// Add(1) 必须在 go 关键字之前调用，
	// 	// 否则可能出现 main 函数的 wg.Wait() 执行时，计数器还是0，导致程序提前退出。
	// 	wg.Add(1) // 在启动一个 goroutine 前，计数器+1
	// 	go func(u string) {
	// 		//defer 保证了即使 checkURL 内部发生 panic，wg.Done() 依然会被执行，
	// 		// 避免 main 函数永久阻塞。
	// 		defer wg.Done() // goroutine 结束时，计数器-1
	// 		result := checkURL(u)
	// 		if result.Error != nil {
	// 			fmt.Printf("URL: %s, 状态: 失败, 错误: %v\n", result.URL, result.Error)
	// 		} else {
	// 			fmt.Printf("URL: %s, 状态码: %d, 延迟: %v\n", result.URL, result.StatusCode, result.Latency)
	// 		}
	// 	}(url) // 注意这里必须把 url 作为参数传入匿名函数，避免闭包问题
	// 	// 因为Goroutine 的启动非常快，但它的执行时机是由 Go 调度器决定的，
	// 	// 可能会有延迟。而 for 循环会继续执行，循环变量 url 的值会不断变化。
	// }

	// wg.Wait() // 阻塞，直到所有 goroutine 都调用了 Done()
	// fmt.Println("检查完成！")

	//  这是出现数据竞争的版本，多个 Goroutine 同时修改 results 切片，每个看到的
	// results长度可能相同，所以可能会重复写入同一个位置，导致结果丢失或乱序。
	var results []CheckResult // 1. 这是被所有 Goroutine 共享的切片
	var wg sync.WaitGroup

	fmt.Println("开始并发检查...")
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result := checkURL(u)

			// 2. 多个 Goroutine 在没有任何同步保护的情况下，同时修改 results
			results = append(results, result) // <--- 数据竞争就发生在这里！
		}(url)
	}

	wg.Wait()
	fmt.Println("检查完成！")
}
