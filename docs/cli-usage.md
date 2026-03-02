# WebDAV CLI 工具使用文档

## 概述

`webdav-cli` 是一个命令行工具，用于执行基本的 WebDAV 操作，如列出目录内容、上传、下载、创建目录和删除文件。

## 安装

```bash
# 从源码构建
cd /path/to/gowebdav
go build -o webdav-cli ./cmd

# 或者使用 go install
go install github.com/AndrewDi/gowebdav/cmd
```

## 基本用法

```bash
webdav-cli -endpoint <endpoint> -username <username> -password <password> -command <command> -remote <remote_path> [-local <local_path>]
```

## 代理配置

### 使用环境变量配置代理

```bash
# 设置 HTTP 代理
export http_proxy=http://127.0.0.0:11111

# 设置 HTTPS 代理（如果需要）
export https_proxy=http://127.0.0.0:11111

# 执行命令
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /
```

## 命令详解

### 1. ls - 列出目录内容

**功能**：列出指定 WebDAV 目录的内容。

**参数**：
- `-remote`：远程目录路径

**示例**：
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /
```

### 2. get - 下载文件

**功能**：从 WebDAV 服务器下载文件到本地。

**参数**：
- `-remote`：远程文件路径
- `-local`：本地保存路径

**示例**：
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command get -remote /document.txt -local ./document.txt
```

### 3. put - 上传文件

**功能**：将本地文件上传到 WebDAV 服务器。

**参数**：
- `-local`：本地文件路径
- `-remote`：远程保存路径

**示例**：
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command put -local ./document.txt -remote /document.txt
```

### 4. mkdir - 创建目录

**功能**：在 WebDAV 服务器上创建目录。

**参数**：
- `-remote`：要创建的目录路径

**示例**：
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command mkdir -remote /new-folder
```

### 5. rm - 删除文件

**功能**：从 WebDAV 服务器删除文件。

**参数**：
- `-remote`：要删除的文件路径

**示例**：
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command rm -remote /document.txt
```

## 完整示例

### 场景：上传文件到服务器并验证

```bash
# 设置代理
export http_proxy=http://127.0.0.0:11111

# 创建测试文件
echo "Hello WebDAV" > test.txt

# 上传文件
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command put -local test.txt -remote /test.txt

# 列出目录内容，确认文件已上传
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /

# 下载文件进行验证
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command get -remote /test.txt -local downloaded.txt

# 查看下载的文件内容
cat downloaded.txt

# 清理文件
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command rm -remote /test.txt
rm test.txt downloaded.txt
```

## 常见问题

### 1. 代理连接失败

**症状**：执行命令时出现连接超时或代理错误。

**解决方法**：
- 确认代理服务器正在运行
- 检查代理地址和端口是否正确
- 验证代理服务器是否允许访问目标 WebDAV 服务器

### 2. 认证失败

**症状**：执行命令时出现 401 Unauthorized 错误。

**解决方法**：
- 确认用户名和密码是否正确
- 验证用户是否有权限访问指定的 WebDAV 资源

### 3. 路径错误

**症状**：执行命令时出现 404 Not Found 错误。

**解决方法**：
- 确认远程路径是否存在
- 对于上传操作，确保目标目录已存在
