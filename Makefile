# Makefile for gowebdav project

# 设置变量
GOPATH ?= $(shell go env GOPATH)
BIN_DIR := $(CURDIR)/build
CMD_DIR := $(CURDIR)/cmd

# 确保bin目录存在
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# 构建命令行工具
.PHONY: build
build:
	@echo "Building webdav-cli..."
	@go build -o $(BIN_DIR)/webdav-cli $(CMD_DIR)/main.go
	@echo "Build completed. Binary is available at $(BIN_DIR)/webdav-cli"

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Clean completed."

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests completed."

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "Dependencies installed."
