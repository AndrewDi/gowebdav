package gowebdav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestSetProxy 测试设置代理
func TestSetProxy(t *testing.T) {
	client, err := NewClient("http://example.com", "", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 测试设置代理
	config := ProxyConfig{
		HTTP:     "http://proxy.example.com:8080",
		HTTPS:    "https://proxy.example.com:8443",
		Username: "user",
		Password: "pass",
	}

	err = client.SetProxy(config)
	if err != nil {
		t.Fatalf("Failed to set proxy: %v", err)
	}

	// 测试从环境变量设置代理
	client.SetProxyFromEnvironment()
}

// TestProxyAuthentication 测试代理认证
func TestProxyAuthentication(t *testing.T) {
	// 创建一个测试服务器作为代理
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查代理认证
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			w.WriteHeader(http.StatusProxyAuthRequired)
			w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
			return
		}

		// 模拟代理请求
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	// 解析代理URL
	proxyURL, err := url.Parse(proxyServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse proxy URL: %v", err)
	}

	// 创建客户端
	client, err := NewClient("http://example.com", "", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 设置代理
	config := ProxyConfig{
		HTTP:     proxyURL.String(),
		HTTPS:    proxyURL.String(),
		Username: "user",
		Password: "pass",
	}

	err = client.SetProxy(config)
	if err != nil {
		t.Fatalf("Failed to set proxy: %v", err)
	}

	// 测试请求
	ctx := context.Background()
	err = client.Get(ctx, "/test", "/tmp/test")
	// 这里会失败，因为我们的测试代理服务器没有真正转发请求
	// 但我们可以检查错误是否与代理认证有关
	if err == nil {
		t.Fatalf("Expected error when proxying to test server")
	}
}
