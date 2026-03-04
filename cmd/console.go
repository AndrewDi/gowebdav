package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/AndrewDi/gowebdav"
	"github.com/AndrewDi/gowebdav/config"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Start interactive console mode",
	Long:  "Start an interactive console for running webdav-cli commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		consoleConfigPath, _ := cmd.Flags().GetString("config")
		consoleEndpoint, _ := cmd.Flags().GetString("endpoint")
		consoleUsername, _ := cmd.Flags().GetString("username")
		consolePassword, _ := cmd.Flags().GetString("password")
		consoleHTTPProxy, _ := cmd.Flags().GetString("http-proxy")
		consoleHTTPSProxy, _ := cmd.Flags().GetString("https-proxy")
		consoleProxyUsername, _ := cmd.Flags().GetString("proxy-username")
		consoleProxyPassword, _ := cmd.Flags().GetString("proxy-password")

		return runConsole(consoleConfigPath, consoleEndpoint, consoleUsername, consolePassword, consoleHTTPProxy, consoleHTTPSProxy, consoleProxyUsername, consoleProxyPassword)
	},
}

var (
	consoleConfigPath    string
	consoleEndpoint      string
	consoleUsername      string
	consolePassword      string
	consoleHTTPProxy     string
	consoleHTTPSProxy    string
	consoleProxyUsername string
	consoleProxyPassword string
	currentPath          string
	webdavClient         *gowebdav.Client
)

func init() {
	currentPath = "/"

	consoleCmd.Flags().StringVar(&consoleConfigPath, "config", "", "Path to config file")
	consoleCmd.Flags().StringVar(&consoleEndpoint, "endpoint", "", "WebDAV server endpoint")
	consoleCmd.Flags().StringVar(&consoleUsername, "username", "", "Username for authentication")
	consoleCmd.Flags().StringVar(&consolePassword, "password", "", "Password for authentication")
	consoleCmd.Flags().StringVar(&consoleHTTPProxy, "http-proxy", "", "HTTP proxy address")
	consoleCmd.Flags().StringVar(&consoleHTTPSProxy, "https-proxy", "", "HTTPS proxy address")
	consoleCmd.Flags().StringVar(&consoleProxyUsername, "proxy-username", "", "Proxy username")
	consoleCmd.Flags().StringVar(&consoleProxyPassword, "proxy-password", "", "Proxy password")

	rootCmd.AddCommand(consoleCmd)
}

func getPrompt() string {
	if currentPath == "/" {
		return "webdav> "
	}
	return fmt.Sprintf("webdav:%s> ", currentPath)
}

func runConsole(cfgPath, endpoint, username, password, httpProxy, httpsProxy, proxyUsername, proxyPassword string) error {
	if err := initWebdavClient(cfgPath, endpoint, username, password, httpProxy, httpsProxy, proxyUsername, proxyPassword); err != nil {
		return err
	}

	fmt.Println("Welcome to WebDAV CLI Console!")
	fmt.Println("Type 'help' for available commands, 'exit' or 'quit' to exit.")
	fmt.Println("Use Tab for auto-completion.")
	fmt.Println()

	cfg := &readline.Config{
		Prompt:       getPrompt(),
		AutoComplete: createCompleter(),
	}
	cfg.Init()

	l, err := readline.NewEx(cfg)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		l.SetPrompt(getPrompt())
		line, err := l.Readline()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println()
				return nil
			}
			fmt.Fprintf(l.Stderr(), "Error: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := executeCommand(line); err != nil {
			if err.Error() == "exit" {
				return nil
			}
			fmt.Fprintf(l.Stderr(), "Error: %v\n", err)
		}
	}
}

func initWebdavClient(cfgPath, endpoint, username, password, httpProxy, httpsProxy, proxyUsername, proxyPassword string) error {
	cfg, err := loadConfig(cfgPath, endpoint, username, password, httpProxy, httpsProxy, proxyUsername, proxyPassword)
	if err != nil {
		return err
	}

	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	webdavClient, err = gowebdav.NewClient(cfg.Endpoint, cfg.Username, cfg.Password)
	if err != nil {
		return err
	}

	webdavClient.SetHTTPClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	if cfg.HTTPProxy != "" || cfg.HTTPSProxy != "" {
		proxyConfig := gowebdav.ProxyConfig{
			HTTP:  cfg.HTTPProxy,
			HTTPS: cfg.HTTPSProxy,
		}
		err := webdavClient.SetProxy(proxyConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadConfig(cfgPath, endpoint, username, password, httpProxy, httpsProxy, proxyUsername, proxyPassword string) (*config.Config, error) {
	configFile := cfgPath
	if configFile == "" {
		configFile = config.GetDefaultConfigPath()
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	if endpoint != "" {
		cfg.Endpoint = endpoint
	}
	if username != "" {
		cfg.Username = username
	}
	if password != "" {
		cfg.Password = password
	}
	if httpProxy != "" {
		cfg.HTTPProxy = httpProxy
	}
	if httpsProxy != "" {
		cfg.HTTPSProxy = httpsProxy
	}
	if proxyUsername != "" {
		cfg.ProxyUsername = proxyUsername
	}
	if proxyPassword != "" {
		cfg.ProxyPassword = proxyPassword
	}

	return cfg, nil
}

func createCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("clear"),
		readline.PcItem("cd",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("pwd"),
		readline.PcItem("ls",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("get",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("put",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("mkdir",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("rm",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("rmdir",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("cat",
			readline.PcItemDynamic(pathCompleter),
		),
		readline.PcItem("vim",
			readline.PcItemDynamic(pathCompleter),
		),
	)
}

func pathCompleter(path string) []string {
	if webdavClient == nil {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." {
		dir = currentPath
	}
	if !strings.HasSuffix(dir, "/") && dir != "/" {
		dir = dir + "/"
	}
	if dir == "/" {
		dir = ""
	}

	var files []gowebdav.FileInfo
	var err error
	if dir == "" {
		files, err = webdavClient.ReadDir(context.Background(), "/")
	} else {
		files, err = webdavClient.ReadDir(context.Background(), dir)
	}
	if err != nil {
		return nil
	}

	var matches []string
	prefix := filepath.Base(path)
	if prefix == "." {
		prefix = ""
	}

	for _, f := range files {
		name := f.Name()
		if prefix == "" || strings.HasPrefix(name, prefix) {
			if f.IsDir() {
				matches = append(matches, name+"/")
			} else {
				matches = append(matches, name)
			}
		}
	}

	return matches
}

func executeCommand(input string) error {
	parts := parseCommand(input)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "help":
		return showHelp()
	case "exit", "quit":
		fmt.Println("Goodbye!")
		return fmt.Errorf("exit")
	case "clear":
		return clearScreen()
	case "cd":
		return executeCd(parts)
	case "pwd":
		return executePwd()
	case "ll":
		args := []string{"ls", "-lrt"}
		if currentPath != "/" {
			args = append(args, currentPath)
		}
		return runCobraCommand(args)
	default:
		return runCobraCommand(parts)
	}
}

func parseCommand(input string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range input {
		if !inQuote && (r == '"' || r == '\'') {
			inQuote = true
			quoteChar = r
		} else if inQuote && r == quoteChar {
			inQuote = false
			quoteChar = 0
		} else if !inQuote && r == ' ' {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func executeCd(args []string) error {
	targetPath := "/"
	if len(args) > 1 {
		targetPath = args[1]
	}

	if targetPath == "" {
		targetPath = "/"
	}

	if targetPath == ".." {
		currentPath = path.Dir(currentPath)
		if currentPath == "." {
			currentPath = "/"
		}
	} else if targetPath == "." {
	} else if strings.HasPrefix(targetPath, "/") {
		currentPath = targetPath
	} else {
		if currentPath == "/" {
			currentPath = "/" + targetPath
		} else {
			currentPath = currentPath + "/" + targetPath
		}
	}

	currentPath = path.Clean(currentPath)
	return nil
}

func executePwd() error {
	fmt.Println(currentPath)
	return nil
}

func showHelp() error {
	fmt.Println("Available commands:")
	fmt.Println("  help           Show this help message")
	fmt.Println("  exit, quit     Exit the console")
	fmt.Println("  clear          Clear the screen")
	fmt.Println("  cd [path]      Change directory (cd, cd .., cd /path)")
	fmt.Println("  pwd            Print working directory")
	fmt.Println()
	fmt.Println("WebDAV Commands:")
	fmt.Println("  ls             List directory contents")
	fmt.Println("  get            Download file from WebDAV server")
	fmt.Println("  put            Upload file to WebDAV server")
	fmt.Println("  mkdir          Create directory on WebDAV server")
	fmt.Println("  rm             Delete file from WebDAV server")
	fmt.Println("  rmdir          Delete directory from WebDAV server")
	fmt.Println("  cat            Display file content")
	fmt.Println("  vim            Edit file with vim")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cd /documents")
	fmt.Println("  ls")
	fmt.Println("  get /remote.txt local.txt")
	fmt.Println("  put local.txt /remote.txt")
	return nil
}

func clearScreen() error {
	fmt.Print("\033[2J\033[H")
	return nil
}

func runCobraCommand(args []string) error {
	adjustedArgs := adjustPathForCommand(args)
	rootCmd.SetArgs(adjustedArgs)
	err := rootCmd.Execute()
	if err != nil && isVimCommand(args) {
		fmt.Println("Editing mode: use vim commands, :wq to save and exit, :q! to discard and exit")
	}
	return err
}

func isVimCommand(args []string) bool {
	return len(args) > 0 && strings.ToLower(args[0]) == "vim"
}

func adjustPathForCommand(args []string) []string {
	if len(args) == 0 {
		return args
	}

	command := strings.ToLower(args[0])

	switch command {
	case "ls":
		args = adjustLsPath(args)
	case "cat", "vim", "mkdir", "rm", "rmdir":
		if len(args) > 1 {
			args[1] = adjustSinglePath(args[1])
		}
	case "get":
		if len(args) > 1 {
			args[1] = adjustSinglePath(args[1])
		}
	case "put":
		if len(args) > 2 {
			args[2] = adjustSinglePath(args[2])
		}
	}

	return args
}

func adjustLsPath(args []string) []string {
	pathIndex := -1

	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			pathIndex = i
			break
		}
	}

	if pathIndex == -1 {
		args = append(args, currentPath)
	} else if args[pathIndex] == "" {
		args[pathIndex] = currentPath
	} else {
		args[pathIndex] = adjustSinglePath(args[pathIndex])
	}

	return args
}

func adjustSinglePath(arg string) string {
	if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "-") || arg == "" {
		return arg
	}
	if currentPath == "/" {
		return "/" + arg
	}
	return currentPath + "/" + arg
}
