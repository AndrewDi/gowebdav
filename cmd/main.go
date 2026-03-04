package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AndrewDi/gowebdav"
	"github.com/AndrewDi/gowebdav/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// 全局变量
var (
	configPath    string
	endpoint      string
	username      string
	password      string
	httpProxy     string
	httpsProxy    string
	proxyUsername string
	proxyPassword string
	lsLong        bool
	lsReverse     bool
	lsTime        bool
)

// 客户端实例
var client *gowebdav.Client

// 根命令
var rootCmd = &cobra.Command{
	Use:   "webdav-cli",
	Short: "WebDAV client",
	Long:  "A command-line client for WebDAV servers",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var err error

		configFile := configPath
		if configFile == "" {
			configFile = config.GetDefaultConfigPath()
		}

		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}

		if endpoint == "" {
			endpoint = cfg.Endpoint
		}
		if username == "" {
			username = cfg.Username
		}
		if password == "" {
			password = cfg.Password
		}
		if httpProxy == "" {
			httpProxy = cfg.HTTPProxy
		}
		if httpsProxy == "" {
			httpsProxy = cfg.HTTPSProxy
		}
		if proxyUsername == "" {
			proxyUsername = cfg.ProxyUsername
		}
		if proxyPassword == "" {
			proxyPassword = cfg.ProxyPassword
		}

		if endpoint == "" {
			log.Fatal("endpoint is required")
		}

		var errClient error
		client, errClient = gowebdav.NewClient(endpoint, username, password)
		if errClient != nil {
			log.Fatalf("Failed to create client: %v", errClient)
		}

		client.SetHTTPClient(&http.Client{
			Timeout: 30 * time.Second,
		})

		if httpProxy != "" || httpsProxy != "" {
			proxyConfig := gowebdav.ProxyConfig{
				HTTP:     httpProxy,
				HTTPS:    httpsProxy,
				Username: proxyUsername,
				Password: proxyPassword,
			}
			err := client.SetProxy(proxyConfig)
			if err != nil {
				log.Fatalf("Failed to set proxy: %v", err)
			}
		}
	},
}

// ls命令
var lsCmd = &cobra.Command{
	Use:   "ls [path] [pattern]",
	Short: "List directory contents",
	Long:  "List the contents of a directory on the WebDAV server, supports pattern matching",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := "/"
		pattern := ""
		if len(args) > 0 {
			if len(args) > 1 {
				remotePath = args[0]
				pattern = args[1]
			} else {
				if strings.ContainsAny(args[0], "*?[]") {
					if strings.HasPrefix(args[0], "/") {
						lastSlash := strings.LastIndex(args[0], "/")
						if lastSlash > 0 {
							remotePath = args[0][:lastSlash]
							pattern = args[0][lastSlash+1:]
						} else {
							remotePath = "/"
							pattern = args[0][1:]
						}
					} else {
						pattern = args[0]
					}
				} else {
					remotePath = args[0]
				}
			}
		}
		ctx := context.Background()
		executeLS(ctx, client, remotePath, lsLong, lsReverse, lsTime, pattern)
	},
}

// get命令
var getCmd = &cobra.Command{
	Use:   "get <remote> <local>",
	Short: "Download file from WebDAV server",
	Long:  "Download a file from the WebDAV server to the local filesystem",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		localPath := args[1]
		ctx := context.Background()
		executeGet(ctx, client, remotePath, localPath)
	},
}

// put命令
var putCmd = &cobra.Command{
	Use:   "put <local> <remote>",
	Short: "Upload file to WebDAV server",
	Long:  "Upload a file from the local filesystem to the WebDAV server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		remotePath := args[1]
		ctx := context.Background()
		executePut(ctx, client, localPath, remotePath)
	},
}

// mkdir命令
var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "Create directory on WebDAV server",
	Long:  "Create a directory on the WebDAV server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		ctx := context.Background()
		executeMkdir(ctx, client, remotePath)
	},
}

// rm命令
var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Delete file from WebDAV server",
	Long:  "Delete a file from the WebDAV server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		ctx := context.Background()
		executeRm(ctx, client, remotePath)
	},
}

// rmdir命令
var rmdirCmd = &cobra.Command{
	Use:   "rmdir <path>",
	Short: "Delete directory from WebDAV server",
	Long:  "Delete a directory from the WebDAV server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		ctx := context.Background()
		executeRmdir(ctx, client, remotePath)
	},
}

// cat命令
var catCmd = &cobra.Command{
	Use:   "cat <path>",
	Short: "Display file content",
	Long:  "Display the content of a file on the WebDAV server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		ctx := context.Background()
		executeCat(ctx, client, remotePath)
	},
}

// vim命令
var vimCmd = &cobra.Command{
	Use:   "vim <path>",
	Short: "Edit file with vim",
	Long:  "Edit a file on the WebDAV server using vim",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		ctx := context.Background()
		executeVim(ctx, client, remotePath)
	},
}

func main() {
	// 添加全局选项
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "WebDAV server endpoint")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Username for authentication")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Password for authentication")
	rootCmd.PersistentFlags().StringVar(&httpProxy, "http-proxy", "", "HTTP proxy address")
	rootCmd.PersistentFlags().StringVar(&httpsProxy, "https-proxy", "", "HTTPS proxy address")
	rootCmd.PersistentFlags().StringVar(&proxyUsername, "proxy-username", "", "Proxy username")
	rootCmd.PersistentFlags().StringVar(&proxyPassword, "proxy-password", "", "Proxy password")

	// 添加ls命令选项
	lsCmd.Flags().BoolVarP(&lsLong, "long", "l", false, "Use long listing format")
	lsCmd.Flags().BoolVarP(&lsReverse, "reverse", "r", false, "Reverse order while sorting")
	lsCmd.Flags().BoolVarP(&lsTime, "time", "t", false, "Sort by time, newest first")

	// 添加子命令
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(rmdirCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(vimCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// executeLS 执行ls命令
func executeLS(ctx context.Context, client *gowebdav.Client, remotePath string, longFormat, reverse, sortByTime bool, pattern string) {
	fmt.Printf("Listing: %s\n", remotePath)

	files, err := client.ReadDir(ctx, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to list: %v", err)
	}

	if pattern != "" {
		regexPattern := pattern
		regexPattern = strings.ReplaceAll(regexPattern, ".", "\\.")
		regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
		regexPattern = "^" + regexPattern + "$"

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			log.Fatalf("❌ Invalid pattern: %v", err)
		}

		var filteredFiles []gowebdav.FileInfo
		for _, file := range files {
			if re.MatchString(file.Name()) {
				filteredFiles = append(filteredFiles, file)
			}
		}
		files = filteredFiles
	}

	currentDirName := path.Base(remotePath)
	if currentDirName == "/" {
		currentDirName = ""
	}

	for i, file := range files {
		if file.IsDir() {
			dirName := strings.TrimSuffix(file.Name(), "/")
			if dirName == currentDirName {
				files[i] = gowebdav.NewFileInfo("./", file.Size(), file.Mode(), file.ModTime(), file.IsDir())
			}
		}
	}

	for i := range files {
		for j := i + 1; j < len(files); j++ {
			if files[i].Name() == "./" {
				continue
			}
			if files[j].Name() == "./" {
				files[i], files[j] = files[j], files[i]
				continue
			}
			if files[i].IsDir() != files[j].IsDir() {
				if files[j].IsDir() {
					files[i], files[j] = files[j], files[i]
				}
			} else {
				if files[i].Name() > files[j].Name() {
					files[i], files[j] = files[j], files[i]
				}
			}
		}
	}

	if sortByTime {
		for i := range files {
			for j := i + 1; j < len(files); j++ {
				if files[i].Name() == "./" || files[j].Name() == "./" {
					continue
				}
				if files[i].IsDir() == files[j].IsDir() {
					if files[i].ModTime().Before(files[j].ModTime()) {
						files[i], files[j] = files[j], files[i]
					}
				}
			}
		}
	}

	if reverse {
		dotIndex := -1
		for i, file := range files {
			if file.Name() == "./" {
				dotIndex = i
				break
			}
		}

		if dotIndex != -1 {
			for i, j := 0, dotIndex-1; i < j; i, j = i+1, j-1 {
				files[i], files[j] = files[j], files[i]
			}
			for i, j := dotIndex+1, len(files)-1; i < j; i, j = i+1, j-1 {
				files[i], files[j] = files[j], files[i]
			}
		} else {
			for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	for i, file := range files {
		if file.Name() == "./" && i != 0 {
			for j := i; j > 0; j-- {
				files[j], files[j-1] = files[j-1], files[j]
			}
			break
		}
	}

	fmt.Println("Contents:")

	if longFormat {
		data := [][]string{}
		data = append(data, []string{"Name", "Size", "Type", "Modified"})

		for _, file := range files {
			name := file.Name()
			var fileType string
			var size string
			if file.IsDir() {
				fileType = "dir"
				size = ""
			} else {
				fileType = "file"
				size = formatSize(file.Size())
			}
			data = append(data, []string{
				name,
				size,
				fileType,
				file.ModTime().Format("2006-01-02 15:04:05"),
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()
	} else {
		for _, file := range files {
			fmt.Printf("%s\n", file.Name())
		}
	}

	fmt.Printf("✓ Listed %d items\n", len(files))
}

// printProgress 打印进度信息
func printProgress(downloaded, total int64) {
	if total <= 0 {
		return
	}

	percentage := float64(downloaded) / float64(total) * 100

	fmt.Fprint(os.Stderr, "\r")

	barLength := 50
	filledLength := int(percentage / 100 * float64(barLength))
	bar := strings.Repeat("=", filledLength) + strings.Repeat(" ", barLength-filledLength)

	fmt.Fprintf(os.Stderr, "[%-50s] %.2f%% (%d/%d)", bar, percentage, downloaded, total)

	if downloaded >= total {
		fmt.Fprint(os.Stderr, "\n")
	}
}

// executeGet 执行get命令
func executeGet(ctx context.Context, client *gowebdav.Client, remotePath, localPath string) {
	fmt.Printf("Downloading %s to %s...\n", remotePath, localPath)

	err := client.GetWithProgress(ctx, remotePath, localPath, printProgress)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}

	fmt.Printf("✓ Downloaded %s to %s\n", remotePath, localPath)
}

// executePut 执行put命令
func executePut(ctx context.Context, client *gowebdav.Client, localPath, remotePath string) {
	fmt.Printf("Uploading %s to %s...\n", localPath, remotePath)

	err := client.PutWithProgress(ctx, localPath, remotePath, printProgress)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Printf("✓ Uploaded %s to %s\n", localPath, remotePath)
}

// executeMkdir 执行mkdir命令
func executeMkdir(ctx context.Context, client *gowebdav.Client, remotePath string) {
	fmt.Printf("Creating directory: %s...\n", remotePath)

	err := client.Mkdir(ctx, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to create directory: %v", err)
	}

	fmt.Printf("✓ Created directory: %s\n", remotePath)
}

// executeRm 执行rm命令
func executeRm(ctx context.Context, client *gowebdav.Client, remotePath string) {
	fmt.Printf("Deleting: %s...\n", remotePath)

	err := client.Delete(ctx, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to delete file: %v", err)
	}

	fmt.Printf("✓ Deleted: %s\n", remotePath)
}

// executeRmdir 执行rmdir命令
func executeRmdir(ctx context.Context, client *gowebdav.Client, remotePath string) {
	fmt.Printf("Deleting directory: %s...\n", remotePath)

	err := client.Rmdir(ctx, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to delete directory: %v", err)
	}

	fmt.Printf("✓ Deleted directory: %s\n", remotePath)
}

// executeCat 执行cat命令
func executeCat(ctx context.Context, client *gowebdav.Client, remotePath string) {
	content, err := client.Read(ctx, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to read file: %v", err)
	}

	switch strings.ToLower(filepath.Ext(remotePath)) {
	case ".json":
		var data any
		if err := json.Unmarshal(content, &data); err == nil {
			formatted, err := json.MarshalIndent(data, "", "  ")
			if err == nil {
				fmt.Println(string(formatted))
				return
			}
		}
	case ".yaml", ".yml":
		var data any
		if err := yaml.Unmarshal(content, &data); err == nil {
			formatted, err := yaml.Marshal(data)
			if err == nil {
				fmt.Println(string(formatted))
				return
			}
		}
	}

	fmt.Println(string(content))
}

// executeVim 执行vim命令
func executeVim(ctx context.Context, client *gowebdav.Client, remotePath string) {
	fmt.Printf("Editing: %s...\n", remotePath)

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "webdav-vim-")
	if err != nil {
		log.Fatalf("❌ Failed to create temporary file: %v", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()

	// 清理临时文件
	defer func() {
		os.Remove(tempFilePath)
	}()

	// 尝试下载远程文件到临时文件
	err = client.Get(ctx, remotePath, tempFilePath)
	if err != nil {
		// 文件不存在，创建一个新的空文件
		fmt.Printf("File does not exist, creating new file...\n")
		tempFile, err = os.Create(tempFilePath)
		if err != nil {
			log.Fatalf("❌ Failed to create temporary file: %v", err)
		}
		tempFile.Close()
	}

	// 调用vim命令编辑临时文件
	cmd := exec.Command("vim", tempFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to edit file: %v", err)
	}

	// 上传修改后的文件回远程服务器
	err = client.Put(ctx, tempFilePath, remotePath)
	if err != nil {
		log.Fatalf("❌ Failed to upload file: %v", err)
	}

	fmt.Printf("✓ Edited: %s\n", remotePath)
}

// formatSize 格式化文件大小，自动转换进制单位
func formatSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	var unit string
	var value float64

	switch {
	case size >= int64(TB):
		value = float64(size) / TB
		unit = "TB"
	case size >= int64(GB):
		value = float64(size) / GB
		unit = "GB"
	case size >= int64(MB):
		value = float64(size) / MB
		unit = "MB"
	case size >= int64(KB):
		value = float64(size) / KB
		unit = "KB"
	default:
		return fmt.Sprintf("%d B", size)
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}
