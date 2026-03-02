# gowebdav

WebDAV client library for Go with command-line interface.

## 中文版本

### 项目简介

gowebdav 是一个基于 Go 语言开发的 WebDAV 客户端库，提供了完整的 WebDAV 协议实现，同时包含一个功能强大的命令行工具 `webdav-cli`。

### 客户端功能

- **目录浏览**：支持列出目录内容，可按时间、名称排序，支持正则匹配
- **文件传输**：支持上传和下载文件，带有进度显示
- **目录操作**：支持创建和删除目录
- **文件操作**：支持删除和查看文件内容
- **远程编辑**：支持使用 vim 编辑远程文件，文件不存在时自动创建
- **代理支持**：支持 HTTP/HTTPS 代理，可配置代理认证
- **配置文件**：支持使用配置文件存储服务器和代理设置

### 安装

```bash
go get github.com/AndrewDi/gowebdav
```

### 快速开始

1. **创建配置文件**：在 `~/.webdav/config.yaml` 中添加服务器和代理设置
2. **列出目录内容**：`webdav-cli ls /`
3. **上传文件**：`webdav-cli put local.txt /remote.txt`
4. **下载文件**：`webdav-cli get /remote.txt local.txt`
5. **编辑文件**：`webdav-cli vim /remote.txt`

## Installation

```bash
go get github.com/AndrewDi/gowebdav
```

## Command-line Interface

The project includes a command-line tool `webdav-cli` for performing basic WebDAV operations.

### Usage

```bash
webdav-cli [global options] <command> [command options] [arguments...]
```

### Global Options

```bash
--config string       Path to config file
--endpoint string     WebDAV server endpoint
--username string     Username for authentication
--password string     Password for authentication
--http-proxy string   HTTP proxy address
--https-proxy string  HTTPS proxy address
--proxy-username string  Proxy username
--proxy-password string  Proxy password
```

### Configuration File

You can use a configuration file to store your WebDAV server and proxy settings, so you don't have to specify them every time you run the command.

#### Default Configuration File Path

The default configuration file path is `~/.webdav/config.yaml`.

#### Configuration File Format

```yaml
# WebDAV客户端配置文件
endpoint: http://example.com:8080
username: your_username
password: your_password
httpProxy: "http://proxy.example.com:8080"
httpsProxy: "http://proxy.example.com:8443"
proxyUsername: ""
proxyPassword: ""
```

#### Using Configuration File

When you run `webdav-cli` without specifying certain parameters, it will automatically read them from the configuration file:

```bash
# This will use the endpoint, username, password, and proxy settings from the configuration file
webdav-cli ls /
```

#### Specifying Custom Configuration File Path

You can use the `--config` parameter to specify a custom configuration file path:

```bash
webdav-cli --config /path/to/config.yaml ls /
```

#### Command-line Parameters Priority

Command-line parameters take precedence over configuration file settings. If you specify a parameter on the command line, it will override the corresponding value in the configuration file:

```bash
# This will use the specified endpoint instead of the one in the configuration file
webdav-cli --endpoint http://example.com:8080 ls /
```

### Proxy Configuration

You can configure proxy settings using environment variables or in the configuration file:

```bash
export http_proxy=http://proxy.example.com:8080
export https_proxy=http://proxy.example.com:8443
webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password ls /
```

You can also specify proxy settings directly on the command line:

```bash
webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password --http-proxy http://proxy.example.com:8080 --https-proxy http://proxy.example.com:8443 ls /
```

### Commands

1. **ls** - List directory contents
   ```bash
   # Basic usage
   webdav-cli ls [path] [pattern]
   
   # With options
   webdav-cli ls -l -r -t [path] [pattern]
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password ls /
   
   # Example with pattern matching
   webdav-cli ls /documents *.txt
   ```
   **Options:**
   - `-l, --long` - Use long listing format
   - `-r, --reverse` - Reverse order while sorting
   - `-t, --time` - Sort by time, newest first

2. **get** - Download file from WebDAV server
   ```bash
   # Basic usage
   webdav-cli get <remote> <local>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password get /file.txt /local/path/file.txt
   ```

3. **put** - Upload file to WebDAV server
   ```bash
   # Basic usage
   webdav-cli put <local> <remote>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password put /local/path/file.txt /file.txt
   ```

4. **mkdir** - Create directory on WebDAV server
   ```bash
   # Basic usage
   webdav-cli mkdir <path>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password mkdir /new/directory
   ```

5. **rm** - Delete file from WebDAV server
   ```bash
   # Basic usage
   webdav-cli rm <path>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password rm /file.txt
   ```

6. **rmdir** - Delete directory from WebDAV server
   ```bash
   # Basic usage
   webdav-cli rmdir <path>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password rmdir /directory
   ```

7. **cat** - Display file content
   ```bash
   # Basic usage
   webdav-cli cat <path>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password cat /file.txt
   ```

8. **vim** - Edit file with vim
   ```bash
   # Basic usage
   webdav-cli vim <path>
   
   # Example with proxy
   export http_proxy=http://proxy.example.com:8080
   webdav-cli --endpoint http://example.com:8080 --username your_username --password your_password vim /file.txt
   ```

## Library Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"

	"gitee.com/AndrewDi/gowebdav"
)

func main() {
	// Create client
	client, err := gowebdav.NewClient("http://example.com:8080", "your_username", "your_password")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Configure proxy
	proxyConfig := gowebdav.ProxyConfig{
		HTTP: "http://proxy.example.com:8080",
		HTTPS: "http://proxy.example.com:8443",
	}
	err = client.SetProxy(proxyConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Alternatively, set proxy from environment variables
	// client.SetProxyFromEnvironment()

	// Create context
	ctx := context.Background()

	// List directory contents
	files, err := client.ReadDir(ctx, "/")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}
```

## API Reference

### Client Methods

- `NewClient(endpoint, username, password string) (*Client, error)` - Create a new WebDAV client
- `SetHTTPClient(client *http.Client)` - Set custom HTTP client
- `Get(ctx context.Context, remotePath, localPath string) error` - Download file
- `Put(ctx context.Context, localPath, remotePath string) error` - Upload file
- `Mkdir(ctx context.Context, remotePath string) error` - Create directory
- `Rmdir(ctx context.Context, remotePath string) error` - Delete directory
- `Delete(ctx context.Context, remotePath string) error` - Delete file
- `ReadDir(ctx context.Context, remotePath string) ([]FileInfo, error)` - List directory contents
