// gateway_test.go
// OFA Center API Gateway Tests (v9.1.0)

package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAPIGateway(t *testing.T) {
	config := DefaultGatewayConfig()
	gw := NewAPIGateway(config)

	if gw == nil {
		t.Fatal("APIGateway should not be nil")
	}
	if gw.router == nil {
		t.Error("Router should not be nil")
	}
	if gw.rateLimiter == nil {
		t.Error("RateLimiter should not be nil (enabled by default)")
	}
	if gw.circuitBreaker == nil {
		t.Error("CircuitBreaker should not be nil (enabled by default)")
	}
	if gw.responseCache == nil {
		t.Error("ResponseCache should not be nil (enabled by default)")
	}
}

func TestAPIGatewayHandle(t *testing.T) {
	gw := NewAPIGateway(DefaultGatewayConfig())

	gw.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status should be 200, got %d", rec.Code)
	}
	if rec.Body.String() != "test response" {
		t.Errorf("Body should be 'test response', got '%s'", rec.Body.String())
	}
}

func TestAPIGatewayNotFound(t *testing.T) {
	gw := NewAPIGateway(DefaultGatewayConfig())

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	rec := httptest.NewRecorder()

	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status should be 404, got %d", rec.Code)
	}
}

func TestTokenBucketRateLimiter(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 5) // 10 req/s, capacity 5

	// 应该允许前5个请求（桶容量）
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 第6个请求应该被拒绝（桶空了）
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if limiter.Allow(req) {
		t.Error("Request 6 should be denied")
	}

	// 重置后应该重新允许
	limiter.Reset()
	if !limiter.Allow(req) {
		t.Error("Request after reset should be allowed")
	}
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(5, 1*time.Minute)

	// 应该允许前5个请求
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 第6个请求应该被拒绝
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if limiter.Allow(req) {
		t.Error("Request 6 should be denied")
	}

	limiter.Reset()
	if !limiter.Allow(req) {
		t.Error("Request after reset should be allowed")
	}
}

func TestCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            5 * time.Second,
		MinRequestCount:    5,
	}
	cb := NewCircuitBreaker(config)

	// 初始状态应该是关闭
	if cb.GetState() != CircuitStateClosed {
		t.Errorf("Initial state should be closed")
	}

	// 记录失败直到达到阈值
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// 状态应该变为开启
	if cb.GetState() != CircuitStateOpen {
		t.Errorf("State should be open after failures")
	}

	// 重置
	cb.Reset()
	if cb.GetState() != CircuitStateClosed {
		t.Errorf("State should be closed after reset")
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:   2,
		SuccessThreshold:   2,
		Timeout:            1 * time.Second,
		MinRequestCount:    3,
	}
	cb := NewCircuitBreaker(config)

	// 触发开启状态
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitStateOpen {
		t.Errorf("State should be open")
	}

	// 等待超时
	time.Sleep(2 * time.Second)

	// 应该进入半开状态
	if cb.GetState() != CircuitStateHalfOpen {
		t.Errorf("State should be half-open after timeout")
	}

	// 记录成功直到恢复
	cb.RecordSuccess(100 * time.Millisecond)
	cb.RecordSuccess(100 * time.Millisecond)

	if cb.GetState() != CircuitStateClosed {
		t.Errorf("State should be closed after successes")
	}
}

func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(100, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// 缓存不存在
	_, found := cache.Get(req)
	if found {
		t.Error("Cache should not exist initially")
	}

	// 设置缓存
	cache.Set(req, []byte("cached response"))

	// 应该能获取到
	body, found := cache.Get(req)
	if !found {
		t.Error("Cache should exist after set")
	}
	if string(body) != "cached response" {
		t.Errorf("Cached body should be 'cached response'")
	}
}

func TestResponseCacheExpiration(t *testing.T) {
	cache := NewResponseCache(100, 100*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	cache.Set(req, []byte("cached"))

	// 立即获取应该存在
	_, found := cache.Get(req)
	if !found {
		t.Error("Cache should exist immediately")
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)

	// 应该过期
	_, found = cache.Get(req)
	if found {
		t.Error("Cache should be expired")
	}
}

func TestResponseCacheMaxSize(t *testing.T) {
	cache := NewResponseCache(3, 5*time.Minute)

	// 添加3个缓存条目
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test"+string(rune('a'+i)), nil)
		cache.Set(req, []byte("data"))
	}

	// 添加第4个条目，应该删除一个旧条目
	req := httptest.NewRequest(http.MethodGet, "/testd", nil)
	cache.Set(req, []byte("data"))

	// 缓存大小应该不超过maxSize
	if len(cache.cache) > cache.maxSize {
		t.Errorf("Cache size %d should not exceed max %d", len(cache.cache), cache.maxSize)
	}
}

func TestMiddlewareChain(t *testing.T) {
	gw := NewAPIGateway(DefaultGatewayConfig())

	order := make([]string, 0)
	gw.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware1")
			next.ServeHTTP(w, r)
		})
	})
	gw.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware2")
			next.ServeHTTP(w, r)
		})
	})

	gw.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)

	// 中间件应该按正确顺序执行
	expected := []string{"middleware1", "middleware2", "handler"}
	if len(order) != len(expected) {
		t.Errorf("Order length should be %d", len(expected))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Order[%d] should be %s, got %s", i, v, order[i])
		}
	}
}

func TestRouter(t *testing.T) {
	router := NewRouter()

	router.HandleFunc("/path1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("path1"))
	})
	router.HandleFunc("/path2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("path2"))
	})

	// 测试path1
	req1 := httptest.NewRequest(http.MethodGet, "/path1", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Body.String() != "path1" {
		t.Errorf("Body should be 'path1'")
	}

	// 测试path2
	req2 := httptest.NewRequest(http.MethodGet, "/path2", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Body.String() != "path2" {
		t.Errorf("Body should be 'path2'")
	}
}

func TestGatewayWithoutRateLimit(t *testing.T) {
	config := DefaultGatewayConfig()
	config.EnableRateLimit = false
	gw := NewAPIGateway(config)

	gw.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// 发送大量请求，不应该被限流
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		gw.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("Request %d should succeed without rate limit", i)
		}
	}
}

func TestGatewayWithoutCache(t *testing.T) {
	config := DefaultGatewayConfig()
	config.EnableCache = false
	gw := NewAPIGateway(config)

	gw.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)

	// 不应该有缓存标记
	if rec.Header().Get("X-Cache") == "HIT" {
		t.Error("Should not have cache HIT header")
	}
}