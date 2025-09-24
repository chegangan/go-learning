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

// 运行go test -race测试模拟数据竞争
func TestConcurrentAppendWithRace(t *testing.T) {
	// 创建一个模拟服务器供所有 goroutine 使用
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 模拟一个 URL 列表
	urls := []string{
		server.URL,
		server.URL,
		server.URL,
		server.URL,
		server.URL,
	}

	var results []CheckResult // 共享的切片
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result := checkURL(u)
			// 在这里并发地 append，这会导致数据竞争
			results = append(results, result)
		}(url)
	}

	wg.Wait()

	// 这个断言可能会失败，因为数据竞争可能导致结果丢失
	if len(results) != len(urls) {
		t.Errorf("期望得到 %d 个结果, 但实际只得到了 %d 个", len(urls), len(results))
	}
}
