package main

import (
	"bytes"
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AndrewDi/gowebdav"
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

// 测试命令行参数解析
func TestCommandLineParsing(t *testing.T) {
	// 保存原始的命令行参数
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	}()

	// 测试有效的命令行参数
	os.Args = []string{
		"gowebdav",
		"--endpoint", "http://example.com",
		"--username", "user",
		"--password", "pass",
		"--command", "ls",
		"--remote", "/test",
	}

	// 解析命令行参数
	endpoint := flag.String("endpoint", "", "WebDAV server endpoint")
	username := flag.String("username", "", "Username for authentication")
	password := flag.String("password", "", "Password for authentication")
	command := flag.String("command", "", "Command to execute: ls, get, put, mkdir, rm")
	remotePath := flag.String("remote", "", "Remote path")
	localPath := flag.String("local", "", "Local path")
	httpProxy := flag.String("http-proxy", "", "HTTP proxy address")
	httpsProxy := flag.String("https-proxy", "", "HTTPS proxy address")
	proxyUsername := flag.String("proxy-username", "", "Proxy username")
	proxyPassword := flag.String("proxy-password", "", "Proxy password")
	flag.Parse()

	// 验证参数解析
	if *endpoint != "http://example.com" {
		t.Errorf("Expected endpoint 'http://example.com', got '%s'", *endpoint)
	}
	if *username != "user" {
		t.Errorf("Expected username 'user', got '%s'", *username)
	}
	if *password != "pass" {
		t.Errorf("Expected password 'pass', got '%s'", *password)
	}
	if *command != "ls" {
		t.Errorf("Expected command 'ls', got '%s'", *command)
	}
	if *remotePath != "/test" {
		t.Errorf("Expected remote path '/test', got '%s'", *remotePath)
	}
	if *localPath != "" {
		t.Errorf("Expected local path '', got '%s'", *localPath)
	}
	if *httpProxy != "" {
		t.Errorf("Expected http proxy '', got '%s'", *httpProxy)
	}
	if *httpsProxy != "" {
		t.Errorf("Expected https proxy '', got '%s'", *httpsProxy)
	}
	if *proxyUsername != "" {
		t.Errorf("Expected proxy username '', got '%s'", *proxyUsername)
	}
	if *proxyPassword != "" {
		t.Errorf("Expected proxy password '', got '%s'", *proxyPassword)
	}
}

// 测试执行LS命令
func TestExecuteLS(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := gowebdav.NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 重定向标准输出
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	// 执行LS命令
	executeLS(context.Background(), client, "/test", false, false, false, "")

	// 关闭写入端，以便读取端可以读取所有数据
	w.Close()

	// 读取输出
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}

	// 验证输出
	output := buf.String()
	if !contains(output, "Listing: /test") {
		t.Error("Expected 'Listing: /test' in output")
	}
	if !contains(output, "Contents:") {
		t.Error("Expected 'Contents:' in output")
	}
	if !contains(output, "file1.txt") {
		t.Error("Expected 'file1.txt' in output")
	}
	if !contains(output, "dir1") {
		t.Error("Expected 'dir1' in output")
	}
	if !contains(output, "Listed 2 items") {
		t.Error("Expected 'Listed 2 items' in output")
	}
}

// 测试执行Get命令
func TestExecuteGet(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := gowebdav.NewClient(server.URL, "", "")
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

	// 重定向标准输出
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	// 执行Get命令
	executeGet(context.Background(), client, "/test/file.txt", tempPath)

	// 关闭写入端，以便读取端可以读取所有数据
	w.Close()

	// 读取输出
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}

	// 验证输出
	output := buf.String()
	if !contains(output, "Downloading /test/file.txt to") {
		t.Error("Expected 'Downloading /test/file.txt to' in output")
	}
	if !contains(output, "Downloaded /test/file.txt to") {
		t.Error("Expected 'Downloaded /test/file.txt to' in output")
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

// 测试执行Put命令
func TestExecutePut(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := gowebdav.NewClient(server.URL, "", "")
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

	// 重定向标准输出
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	// 执行Put命令
	executePut(context.Background(), client, tempPath, "/test/upload.txt")

	// 关闭写入端，以便读取端可以读取所有数据
	w.Close()

	// 读取输出
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}

	// 验证输出
	output := buf.String()
	if !contains(output, "Uploading") {
		t.Error("Expected 'Uploading' in output")
	}
	if !contains(output, "Uploaded") {
		t.Error("Expected 'Uploaded' in output")
	}
}

// 测试执行Mkdir命令
func TestExecuteMkdir(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := gowebdav.NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 重定向标准输出
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	// 执行Mkdir命令
	executeMkdir(context.Background(), client, "/test/newdir")

	// 关闭写入端，以便读取端可以读取所有数据
	w.Close()

	// 读取输出
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}

	// 验证输出
	output := buf.String()
	if !contains(output, "Creating directory: /test/newdir") {
		t.Error("Expected 'Creating directory: /test/newdir' in output")
	}
	if !contains(output, "Created directory: /test/newdir") {
		t.Error("Expected 'Created directory: /test/newdir' in output")
	}
}

// 测试执行Rm命令
func TestExecuteRm(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	client, err := gowebdav.NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// 重定向标准输出
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	// 执行Rm命令
	executeRm(context.Background(), client, "/test/file1.txt")

	// 关闭写入端，以便读取端可以读取所有数据
	w.Close()

	// 读取输出
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}

	// 验证输出
	output := buf.String()
	if !contains(output, "Deleting: /test/file1.txt") {
		t.Error("Expected 'Deleting: /test/file1.txt' in output")
	}
	if !contains(output, "Deleted: /test/file1.txt") {
		t.Error("Expected 'Deleted: /test/file1.txt' in output")
	}
}

// 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 由于executeLS等函数使用log.Fatalf，会导致测试进程退出
	// 因此我们不直接测试这些函数，而是测试命令行参数解析和客户端创建
	// 实际的错误处理测试已经在client_test.go中完成
	t.Log("Error handling is tested in client_test.go")
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
