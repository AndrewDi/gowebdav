package gowebdav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// 模拟WebDAV服务器
func mockWebDAVServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// 模拟Get请求
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test content"))
		case http.MethodPut:
			// 模拟Put请求
			w.WriteHeader(http.StatusCreated)
		case "MKCOL":
			// 模拟Mkdir请求
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			// 模拟Delete请求
			w.WriteHeader(http.StatusNoContent)
		case "PROPFIND":
			// 模拟ReadDir请求
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusMultiStatus)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<multistatus xmlns="DAV:">
  <response>
    <href>/test/file1.txt</href>
    <propstat>
      <prop>
        <getcontentlength>10</getcontentlength>
        <getlastmodified>Sun, 06 Nov 1994 08:49:37 GMT</getlastmodified>
        <resourcetype/>
        <displayname>file1.txt</displayname>
      </prop>
      <status>HTTP/1.1 200 OK</status>
    </propstat>
  </response>
  <response>
    <href>/test/dir1/</href>
    <propstat>
      <prop>
        <getcontentlength>0</getcontentlength>
        <getlastmodified>Sun, 06 Nov 1994 08:49:37 GMT</getlastmodified>
        <resourcetype>
          <collection/>
        </resourcetype>
        <displayname>dir1</displayname>
      </prop>
      <status>HTTP/1.1 200 OK</status>
    </propstat>
  </response>
</multistatus>`))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
}

// 模拟错误的WebDAV服务器
func mockErrorWebDAVServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
}

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name     string
		endpoint string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid endpoint",
			endpoint: "http://example.com/webdav",
			username: "user",
			password: "pass",
			wantErr:  false,
		},
		{
			name:     "Invalid endpoint",
			endpoint: "invalid-url",
			username: "user",
			password: "pass",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.endpoint, tc.username, tc.password)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client")
				}
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("os.CreateTemp() error = %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// 测试Get方法
	err = client.Get(context.Background(), "/test/file.txt", tempPath)
	if err != nil {
		t.Errorf("Client.Get() error = %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(tempPath)
	if err != nil {
		t.Errorf("os.ReadFile() error = %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", string(content))
	}
}

func TestClient_Put(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("os.CreateTemp() error = %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Write([]byte("test upload"))
	tempFile.Close()
	defer os.Remove(tempPath)

	// 测试Put方法
	err = client.Put(context.Background(), tempPath, "/test/upload.txt")
	if err != nil {
		t.Errorf("Client.Put() error = %v", err)
	}
}

func TestClient_Mkdir(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试Mkdir方法
	err = client.Mkdir(context.Background(), "/test/newdir")
	if err != nil {
		t.Errorf("Client.Mkdir() error = %v", err)
	}
}

func TestClient_Rmdir(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试Rmdir方法
	err = client.Rmdir(context.Background(), "/test/dir1")
	if err != nil {
		t.Errorf("Client.Rmdir() error = %v", err)
	}
}

func TestClient_Delete(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试Delete方法
	err = client.Delete(context.Background(), "/test/file1.txt")
	if err != nil {
		t.Errorf("Client.Delete() error = %v", err)
	}
}

func TestClient_ReadDir(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试ReadDir方法
	files, err := client.ReadDir(context.Background(), "/test")
	if err != nil {
		t.Errorf("Client.ReadDir() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// 验证文件信息
	fileMap := make(map[string]FileInfo)
	for _, file := range files {
		fileMap[file.Name()] = file
	}

	if _, ok := fileMap["file1.txt"]; !ok {
		t.Error("Expected file1.txt in directory listing")
	}

	if _, ok := fileMap["dir1/"]; !ok {
		t.Error("Expected dir1/ in directory listing")
	}

	if fileMap["file1.txt"].IsDir() {
		t.Error("file1.txt should not be a directory")
	}

	if !fileMap["dir1/"].IsDir() {
		t.Error("dir1/ should be a directory")
	}
}

func TestClient_Copy(t *testing.T) {
	client, err := NewClient("http://example.com", "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试Copy方法（未实现）
	err = client.Copy(context.Background(), "/source", "/dest")
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got %v", err)
	}
}

func TestClient_Move(t *testing.T) {
	client, err := NewClient("http://example.com", "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试Move方法（未实现）
	err = client.Move(context.Background(), "/source", "/dest")
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got %v", err)
	}
}

func TestClient_SetProxy(t *testing.T) {
	client, err := NewClient("http://example.com", "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试SetProxy方法
	config := ProxyConfig{
		HTTP:     "http://proxy:8080",
		HTTPS:    "https://proxy:8443",
		Username: "proxyuser",
		Password: "proxypass",
	}

	err = client.SetProxy(config)
	if err != nil {
		t.Errorf("Client.SetProxy() error = %v", err)
	}

	// 测试SetProxyFromEnvironment方法
	client.SetProxyFromEnvironment()
}

func TestClient_ErrorHandling(t *testing.T) {
	server := mockErrorWebDAVServer()
	defer server.Close()

	client, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 测试错误处理
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("os.CreateTemp() error = %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// 测试Get方法错误处理
	err = client.Get(context.Background(), "/test/file.txt", tempPath)
	if err == nil {
		t.Error("Expected error from Get method")
	}

	// 测试Put方法错误处理
	err = client.Put(context.Background(), tempPath, "/test/upload.txt")
	if err == nil {
		t.Error("Expected error from Put method")
	}

	// 测试Mkdir方法错误处理
	err = client.Mkdir(context.Background(), "/test/newdir")
	if err == nil {
		t.Error("Expected error from Mkdir method")
	}

	// 测试Rmdir方法错误处理
	err = client.Rmdir(context.Background(), "/test/dir1")
	if err == nil {
		t.Error("Expected error from Rmdir method")
	}

	// 测试Delete方法错误处理
	err = client.Delete(context.Background(), "/test/file1.txt")
	if err == nil {
		t.Error("Expected error from Delete method")
	}

	// 测试ReadDir方法错误处理
	_, err = client.ReadDir(context.Background(), "/test")
	if err == nil {
		t.Error("Expected error from ReadDir method")
	}
}
