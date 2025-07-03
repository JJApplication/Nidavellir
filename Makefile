# Nidavellir 配置中心 Makefile

.PHONY: build run clean test proto deps help

# 默认目标
all: build

# 构建项目
build:
	@echo "Building Nidavellir..."
	go build -o bin/nidavellir main.go

# 运行项目
run:
	@echo "Running Nidavellir..."
	go run main.go

# 清理构建文件
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf api/proto/config/*.pb.go

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 生成protobuf文件
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/config.proto

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
vet:
	@echo "Running go vet..."
	go vet ./...

# 安装工具
tools:
	@echo "Installing tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t nidavellir:latest .

# Docker运行
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 nidavellir:latest

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build       - Build the project"
	@echo "  run         - Run the project"
	@echo "  clean       - Clean build files"
	@echo "  test        - Run tests"
	@echo "  proto       - Generate protobuf files"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  tools       - Install required tools"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run Docker container"
	@echo "  help        - Show this help"