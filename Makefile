.PHONY: test coverage lint mod help

# 输出目录
OUT_DIR=out
# 版本号
VERSION=v0.1.0
# Git 提交哈希
COMMIT=$(shell git rev-parse --short HEAD)
# 构建时间
BUILD_TIME=$(shell date '+%Y-%m-%d %H:%M:%S')

# 默认目标
.DEFAULT_GOAL := help

help:
	@echo "使用方法:"
	@echo "  make <目标>"
	@echo ""
	@echo "目标:"
	@echo "  test      运行测试"
	@echo "  coverage  生成测试覆盖率报告"
	@echo "  lint      运行代码检查"
	@echo "  mod       更新 Go 模块依赖"
	@echo "  clean     清理输出目录"
	@echo "  help      显示帮助信息"

test:
	@echo "运行测试..."
	@go test -v ./...

coverage:
	@echo "生成测试覆盖率报告..."
	@mkdir -p $(OUT_DIR)
	@go test -coverprofile=$(OUT_DIR)/coverage.out ./...
	@go tool cover -html=$(OUT_DIR)/coverage.out -o $(OUT_DIR)/coverage.html

lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "请先安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

mod:
	@echo "更新依赖..."
	@go mod tidy
	@go mod verify

clean:
	@echo "清理输出目录..."
	@rm -rf $(OUT_DIR) 