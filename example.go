package gowebdav

import (
	"context"
	"fmt"
)

// ExampleClient 展示如何使用WebDAV客户端
func ExampleClient() {
	// 创建一个新的WebDAV客户端
	client, err := NewClient("http://192.168.1.5:5005", "webcli", "@Andrewtomy9")
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}

	// 配置代理（使用用户提供的代理配置）
	proxyConfig := ProxyConfig{
		HTTP:  "http://127.0.0.0:11111",
		HTTPS: "http://127.0.0.0:11111",
	}
	err = client.SetProxy(proxyConfig)
	if err != nil {
		fmt.Printf("设置代理失败: %v\n", err)
		return
	}

	// 配置带认证的代理
	// proxyConfigWithAuth := ProxyConfig{
	// 	HTTP:      "http://proxy.example.com:8080",
	// 	HTTPS:     "http://proxy.example.com:8443",
	// 	Username:  "proxyuser",
	// 	Password:  "proxypass",
	// }
	// err = client.SetProxy(proxyConfigWithAuth)
	// if err != nil {
	// 	fmt.Printf("设置代理失败: %v\n", err)
	// 	return
	// }

	// 或者从环境变量中设置代理
	// client.SetProxyFromEnvironment()

	// 创建目录
	err = client.Mkdir(context.Background(), "/test")
	if err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
	}

	// 上传文件
	err = client.Put(context.Background(), "/local/file.txt", "/test/file.txt")
	if err != nil {
		fmt.Printf("上传文件失败: %v\n", err)
	}

	// 下载文件
	err = client.Get(context.Background(), "/test/file.txt", "/local/downloaded.txt")
	if err != nil {
		fmt.Printf("下载文件失败: %v\n", err)
	}

	// 读取目录内容
	files, err := client.ReadDir(context.Background(), "/test")
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
	} else {
		fmt.Println("目录内容:")
		for _, file := range files {
			fmt.Printf("%s (大小: %d, 是否目录: %t)\n", file.Name(), file.Size(), file.IsDir())
		}
	}

	// 移动文件
	err = client.Move(context.Background(), "/test/file.txt", "/test/file_moved.txt")
	if err != nil {
		fmt.Printf("移动文件失败: %v\n", err)
	}

	// 复制文件
	err = client.Copy(context.Background(), "/test/file_moved.txt", "/test/file_copy.txt")
	if err != nil {
		fmt.Printf("复制文件失败: %v\n", err)
	}

	// 删除文件
	err = client.Delete(context.Background(), "/test/file_moved.txt")
	if err != nil {
		fmt.Printf("删除文件失败: %v\n", err)
	}

	// 删除目录
	err = client.Rmdir(context.Background(), "/test")
	if err != nil {
		fmt.Printf("删除目录失败: %v\n", err)
	}
}
