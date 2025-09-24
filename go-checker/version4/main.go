package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
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

// worker 函数从 jobs channel 接收任务，并将结果发送到 results channel
func worker(id int, jobs <-chan string, results chan<- CheckResult) {
	for url := range jobs {
		fmt.Printf("Worker %d 开始处理 %s\n", id, url)
		result := checkURL(url)
		results <- result
	}
}

// 把任务分发、工作、结果收集三块分开
func main() {
	// 1. 使用 flag 包接收命令行传入的文件名
	filePath := flag.String("file", "urls.txt", "包含URL列表的文件路径")
	concurrency := flag.Int("c", 10, "并发的 worker 数量")
	flag.Parse() // 注意 flag.Parse() 只调用一次

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

	// 创建任务 channel 和结果 channel
	jobs := make(chan string, len(urls))
	results := make(chan CheckResult, len(urls))

	// 启动指定数量的 worker
	for w := 1; w <= *concurrency; w++ {
		go worker(w, jobs, results)
	}

	// 将所有 URL 发送到任务 channel
	for _, url := range urls {
		jobs <- url
	}
	close(jobs) // 发送完所有任务后，关闭 jobs channel

	var allResults []CheckResult
	// 收集所有结果
	for a := 1; a <= len(urls); a++ {
		result := <-results
		allResults = append(allResults, result)
	}

	// 使用 tabwriter 格式化输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "URL\tStatusCode\tLatency\tError\t")
	fmt.Fprintln(w, "---\t----------\t-------\t-----\t")

	var successCount, failCount int
	var totalLatency time.Duration

	for _, res := range allResults {
		if res.Error != nil {
			fmt.Fprintf(w, "%s\tN/A\tN/A\t%v\t\n", res.URL, res.Error)
			failCount++
		} else {
			fmt.Fprintf(w, "%s\t%d\t%v\t%s\t\n", res.URL, res.StatusCode, res.Latency, "N/A")
			successCount++
			totalLatency += res.Latency
		}
	}
	w.Flush() // 不要忘记 Flush

	// 打印统计信息
	fmt.Println("\n--- 统计信息 ---")
	fmt.Printf("总计URL数量: %d\n", len(allResults))
	fmt.Printf("成功数量: %d\n", successCount)
	fmt.Printf("失败数量: %d\n", failCount)
	if successCount > 0 {
		fmt.Printf("平均延迟: %v\n", totalLatency/time.Duration(successCount))
	}
	fmt.Println("-----------------")
}
