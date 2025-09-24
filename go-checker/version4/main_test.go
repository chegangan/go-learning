package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCheckURL(t *testing.T) {
	// 创建一个模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟一个正常的响应
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := checkURL(server.URL)

	if result.Error != nil {
		t.Errorf("期望没有错误，但得到了: %v", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d, 但得到了 %d", http.StatusOK, result.StatusCode)
	}
	if result.Latency <= 0 {
		t.Error("期望延迟大于0")
	}

	// 测试一个无效的 URL
	invalidResult := checkURL("[http://invalid-url-that-will-fail.com](http://invalid-url-that-will-fail.com)")
	if invalidResult.Error == nil {
		t.Error("期望得到一个错误，但没有得到")
	}
}

// 辅助函数：生成包含 n 个URL的切片
func generateURLs(n int) []string {
	// 基础URL列表，用于循环生成
	baseUrls := []string{
		"https://www.google.com",
		"https://www.baidu.com",
		"https://github.com",
		"https://www.youtube.com",
		"https://www.amazon.com",
	}

	urls := make([]string, 0, n)
	for i := 0; i < n; i++ {
		// 使用取模运算循环从基础列表中选取URL
		urls = append(urls, baseUrls[i%len(baseUrls)])
	}
	return urls
}

// 测试串行速度
func BenchmarkSequentialCheck(b *testing.B) {
	// 轻松修改这里的数字来增减URL数量
	urls := generateURLs(100)

	// b.ResetTimer() // 重置计时器，忽略上面生成URL的时间

	for i := 0; i < b.N; i++ {
		for _, url := range urls {
			checkURL(url)
		}
	}
}

// 测试并发速度
func BenchmarkConcurrentCheck(b *testing.B) {
	// 轻松修改这里的数字来增减URL数量
	urls := generateURLs(100)

	// b.ResetTimer() // 重置计时器，忽略上面生成URL的时间

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				checkURL(u)
			}(url)
		}
		wg.Wait()
	}
}
