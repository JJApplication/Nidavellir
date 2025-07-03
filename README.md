# Nidavellir 配置中心

Nidavellir 是一个基于 etcd 的分布式配置中心，提供 HTTP 和 gRPC 两种接口方式，支持多服务配置管理和实时配置监听。

## 功能特性

- 🚀 **高性能**: 基于 etcd 存储，支持高并发读写
- 🔄 **实时同步**: 支持配置变更实时推送
- 🌐 **多协议**: 同时支持 HTTP RESTful API 和 gRPC 接口
- 🏢 **多服务**: 基于服务名称进行配置隔离
- 📊 **监控友好**: 内置健康检查和日志记录
- 🐳 **容器化**: 支持 Docker 和 Docker Compose 部署
- 🔧 **易于使用**: 简单的 API 设计，易于集成

## 快速开始

### 前置要求

- Go 1.21+
- etcd 3.5+
- Protocol Buffers 编译器 (可选，用于重新生成 protobuf 文件)

### 本地开发

1. **安装依赖**
```bash
make deps
```

2. **启动 etcd**
```bash
# 使用 Docker
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.10 \
  etcd --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://localhost:2379
```

3. **运行项目**
```bash
make run
```

### Docker Compose 部署

```bash
# 启动所有服务（包括 etcd）
docker-compose up -d

# 查看日志
docker-compose logs -f nidavellir

# 停止服务
docker-compose down
```

## API 文档

### HTTP API

基础 URL: `http://localhost:8080/api/v1`

#### 健康检查
```http
GET /health
```

#### 配置管理

**设置配置**
```http
PUT /configs/{service}/{key}
Content-Type: application/json

{
  "value": "配置值",
  "description": "配置描述"
}
```

**获取配置**
```http
GET /configs/{service}/{key}
```

**获取服务所有配置**
```http
GET /configs/{service}
```

**删除配置**
```http
DELETE /configs/{service}/{key}
```

**删除服务所有配置**
```http
DELETE /configs/{service}
```

**列出所有服务**
```http
GET /services
```

### gRPC API

gRPC 服务运行在 `localhost:9090`，详细的 API 定义请参考 `api/proto/config.proto`。

## 使用示例

### HTTP API 示例

```bash
# 设置配置
curl -X PUT http://localhost:8080/api/v1/configs/user-service/database_url \
  -H "Content-Type: application/json" \
  -d '{
    "value": "mysql://user:pass@localhost:3306/userdb",
    "description": "用户服务数据库连接地址"
  }'

# 获取配置
curl http://localhost:8080/api/v1/configs/user-service/database_url

# 获取服务所有配置
curl http://localhost:8080/api/v1/configs/user-service

# 列出所有服务
curl http://localhost:8080/api/v1/services
```

## 项目结构

```
Nidavellir/
├── api/
│   └── proto/           # Protocol Buffers 定义
├── configs/             # 配置文件
├── internal/
│   ├── config/          # 配置管理
│   ├── etcd/           # etcd 客户端和服务
│   ├── grpc/           # gRPC 服务器
│   └── http/           # HTTP 服务器
├── pkg/
│   └── logger/         # 日志工具
├── main.go             # 程序入口
├── go.mod              # Go 模块定义
├── Makefile            # 构建脚本
├── Dockerfile          # Docker 镜像定义
└── docker-compose.yml  # Docker Compose 配置
```

## 配置文件

配置文件位于 `configs/config.toml`：

```toml
# Nidavellir 配置中心配置文件

# HTTP服务器配置
[http]
host = "0.0.0.0"
port = 8080

# gRPC服务器配置
[grpc]
host = "0.0.0.0"
port = 9090

# etcd配置
[etcd]
endpoints = ["localhost:2379"]
dial_timeout = 5
username = ""
password = ""

# 日志配置
[log]
level = "info"
format = "json"
```

## 常用命令

```bash
# 构建项目
make build

# 运行项目
make run

# 运行测试
make test

# 生成 protobuf 文件
make proto

# 格式化代码
make fmt

# 代码检查
make vet

# 查看帮助
make help
```
Nidavellir
