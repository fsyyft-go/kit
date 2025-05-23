.PHONY: test coverage lint mod help download verify

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
	@echo "  test      运行测试和构建示例"
	@echo "  coverage  生成测试覆盖率报告"
	@echo "  lint      运行代码检查"
	@echo "  mod       更新 Go 模块依赖"
	@echo "  clean     清理输出目录"
	@echo "  help      显示帮助信息"

test:
	@echo "构建示例程序..."
	@chmod +x example/cache/build.sh \
		example/config/version/build.sh \
		example/database/sql/gorm/build.sh \
		example/database/sql/mysql/build.sh \
		example/kratos/config/build.sh \
		example/log/build.sh \
		example/net/message/build.sh \
		example/runtime/goroutine/build.sh
	
	@echo "\n===================================================="
	@echo "构建并运行缓存示例..."
	@example/cache/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行版本信息示例..."
	@example/config/version/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行 kratos config 示例..."
	@example/kratos/config/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行 GORM 示例..."
	@example/database/sql/gorm/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行 MySQL 示例..."
	@example/database/sql/mysql/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行日志示例..."
	@example/log/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行 net message 示例..."
	@example/net/message/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "构建并运行获取 goroutine id 示例..."
	@example/runtime/goroutine/build.sh
	@echo "====================================================\n"
	
	@echo "===================================================="
	@echo "运行单元测试..."
	@go test -v -race ./...
	@echo "====================================================\n"

coverage:
	@echo "生成测试覆盖率报告..."
	@mkdir -p $(OUT_DIR)
	@go test -v -race -coverprofile=$(OUT_DIR)/coverage.txt -covermode=atomic ./...
	@go tool cover -html=$(OUT_DIR)/coverage.txt -o $(OUT_DIR)/coverage.html

lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "请先安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

download:
	@echo "下载依赖..."
	@go mod download

verify:
	@echo "验证依赖..."
	@go mod verify

mod:
	@echo "更新依赖..."
	@go mod tidy
	@go mod verify

clean:
	@echo "清理输出目录..."
	@rm -rf $(OUT_DIR) bin/ 