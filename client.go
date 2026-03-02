package gowebdav

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// WebDAVError 表示WebDAV操作的错误
// 包含HTTP状态码和详细错误信息
type WebDAVError struct {
	StatusCode int    // HTTP状态码
	Message    string // 错误信息
	Operation  string // 操作类型（如get, put, mkdir等）
	Path       string // 操作的路径
}

// Error 实现error接口
func (e *WebDAVError) Error() string {
	return fmt.Sprintf("webdav error: %s %s (status code: %d): %s", e.Operation, e.Path, e.StatusCode, e.Message)
}

// Client 表示WebDAV客户端
type Client struct {
	http     *http.Client
	endpoint *url.URL
	username string
	password string
	rootPath string
}

// NewClient 创建一个新的WebDAV客户端
func NewClient(endpoint, username, password string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// 检查是否有scheme
	if u.Scheme == "" {
		return nil, fmt.Errorf("invalid endpoint: missing scheme")
	}

	// 确保rootPath以/开头和结尾
	rootPath := u.Path
	if !strings.HasPrefix(rootPath, "/") {
		rootPath = "/" + rootPath
	}
	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}
	u.Path = ""

	return &Client{
		http:     http.DefaultClient,
		endpoint: u,
		username: username,
		password: password,
		rootPath: rootPath,
	}, nil
}

// SetHTTPClient 设置自定义的HTTP客户端
func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

// setAuth 设置请求的认证信息
func (c *Client) setAuth(req *http.Request) {
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

// buildPath 构建完整的路径
func (c *Client) buildPath(p string) string {
	// 确保路径以/开头
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// 组合rootPath和p，使用path.Join避免重复的/
	return path.Join(c.rootPath, p[1:])
}

// buildURL 构建完整的URL
func (c *Client) buildURL(p string) string {
	u := *c.endpoint
	path := c.buildPath(p)

	// 对路径的每个部分进行URL编码，确保特殊字符（如中文字符）能够正确处理
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	encodedPath := strings.Join(parts, "/")

	// 直接构建URL，避免url包的自动编码
	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, encodedPath)
}

// ProgressCallback 定义进度回调函数类型
type ProgressCallback func(downloaded, total int64)

// Get 下载文件
func (c *Client) Get(ctx context.Context, remotePath, localPath string) error {
	return c.GetWithProgress(ctx, remotePath, localPath, nil)
}

// Read 读取文件内容到内存
func (c *Client) Read(ctx context.Context, remotePath string) ([]byte, error) {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "read",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "read",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "read",
			Path:       remotePath,
		}
	}

	// 读取内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "read",
			Path:       remotePath,
		}
	}

	return content, nil
}

// GetWithProgress 下载文件并显示进度
func (c *Client) GetWithProgress(ctx context.Context, remotePath, localPath string, callback ProgressCallback) error {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "get",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "get",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "get",
			Path:       remotePath,
		}
	}

	// 创建本地文件
	file, err := os.Create(localPath)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "get",
			Path:       localPath,
		}
	}
	defer file.Close()

	// 获取文件大小
	total := resp.ContentLength

	// 复制内容并显示进度
	if callback != nil {
		// 创建带进度的reader
		r := &progressReader{
			r:        resp.Body,
			total:    total,
			read:     0,
			callback: callback,
		}
		_, err = io.Copy(file, r)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "get",
			Path:       remotePath,
		}
	}
	return nil
}

// progressReader 包装io.Reader以提供进度回调
type progressReader struct {
	r        io.Reader
	total    int64
	read     int64
	callback ProgressCallback
}

// Read 实现io.Reader接口
func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	pr.read += int64(n)
	if pr.callback != nil {
		pr.callback(pr.read, pr.total)
	}
	return
}

// Put 上传文件
func (c *Client) Put(ctx context.Context, localPath, remotePath string) error {
	return c.PutWithProgress(ctx, localPath, remotePath, nil)
}

// PutWithProgress 上传文件并显示进度
func (c *Client) PutWithProgress(ctx context.Context, localPath, remotePath string, callback ProgressCallback) error {
	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "put",
			Path:       localPath,
		}
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "put",
			Path:       localPath,
		}
	}
	total := fileInfo.Size()

	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建带进度的reader
	var body io.Reader
	if callback != nil {
		body = &progressReader{
			r:        file,
			total:    total,
			read:     0,
			callback: callback,
		}
	} else {
		body = file
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "put",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "put",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "put",
			Path:       remotePath,
		}
	}

	return nil
}

// Mkdir 创建目录
func (c *Client) Mkdir(ctx context.Context, remotePath string) error {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "MKCOL", url, nil)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "mkdir",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "mkdir",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated {
		return &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "mkdir",
			Path:       remotePath,
		}
	}

	return nil
}

// Rmdir 删除目录
func (c *Client) Rmdir(ctx context.Context, remotePath string) error {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "rmdir",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "rmdir",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusNoContent {
		return &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "rmdir",
			Path:       remotePath,
		}
	}

	return nil
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, remotePath string) error {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "delete",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "delete",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusNoContent {
		return &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "delete",
			Path:       remotePath,
		}
	}

	return nil
}

// Copy 复制文件或目录
func (c *Client) Copy(ctx context.Context, sourcePath, destinationPath string) error {
	// 实现复制逻辑
	return fmt.Errorf("not implemented")
}

// Move 移动文件或目录
func (c *Client) Move(ctx context.Context, sourcePath, destinationPath string) error {
	// 实现移动逻辑
	return fmt.Errorf("not implemented")
}

// ReadDir 读取目录内容
func (c *Client) ReadDir(ctx context.Context, remotePath string) ([]FileInfo, error) {
	// 构建完整的URL
	url := c.buildURL(remotePath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", url, nil)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "readdir",
			Path:       remotePath,
		}
	}

	// 设置认证信息
	c.setAuth(req)

	// 设置深度头
	req.Header.Set("Depth", "1")

	// 发送请求
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "readdir",
			Path:       remotePath,
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusMultiStatus {
		return nil, &WebDAVError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Operation:  "readdir",
			Path:       remotePath,
		}
	}

	// 解析XML响应
	files, err := parsePropfindResponse(resp.Body)
	if err != nil {
		return nil, &WebDAVError{
			StatusCode: 0,
			Message:    err.Error(),
			Operation:  "readdir",
			Path:       remotePath,
		}
	}

	return files, nil
}

// 定义XML解析结构体
type propfindResponse struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []response `xml:"response"`
}

type response struct {
	Href     string   `xml:"href"`
	Propstat propstat `xml:"propstat"`
}

type propstat struct {
	Prop   prop   `xml:"prop"`
	Status string `xml:"status"`
}

type prop struct {
	GetContentLength int64        `xml:"getcontentlength"`
	GetLastModified  string       `xml:"getlastmodified"`
	ResourceType     resourceType `xml:"resourcetype"`
	DisplayName      string       `xml:"displayname"`
}

type resourceType struct {
	Collection string `xml:"collection"`
}

// parsePropfindResponse 解析PROPFIND响应
func parsePropfindResponse(body io.Reader) ([]FileInfo, error) {
	var resp propfindResponse
	if err := xml.NewDecoder(body).Decode(&resp); err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, r := range resp.Responses {
		// 跳过当前目录的特殊情况，但保留文件本身的信息
		if strings.HasSuffix(r.Href, "/.") {
			continue
		}

		// 提取文件名并解码
		name := path.Base(r.Href)
		if name == "/" {
			// 如果是根目录，跳过
			continue
		}

		// 对文件名进行URL解码，确保中文字符能够正确显示
		decodedName, err := url.PathUnescape(name)
		if err == nil {
			name = decodedName
		}

		// 解析修改时间
		modTime, err := time.Parse(time.RFC1123, r.Propstat.Prop.GetLastModified)
		if err != nil {
			modTime = time.Now()
		}

		// 检查是否为目录
		isDir := strings.HasSuffix(r.Href, "/")

		// 如果是目录，在名称末尾添加斜杠
		if isDir {
			name += "/"
		}

		// 创建FileInfo
		file := NewFileInfo(name, r.Propstat.Prop.GetContentLength, 0, modTime, isDir)
		files = append(files, file)
	}

	return files, nil
}
