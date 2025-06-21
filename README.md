# Go Redis Server

一个用 Go 语言实现的简单 Redis 服务器。

## 功能特性

- TCP ServerSocket 实现
- 支持多客户端并发连接
- 基础的 Redis 命令支持 (PING, ECHO, INFO, QUIT, SET, GET)
- 可配置的监听地址和端口
- 线程安全的内存存储
- 完整的 RESP (Redis Serialization Protocol) 协议支持

## 使用方法

### 启动服务器

```bash
# 使用默认配置 (127.0.0.1:6379)
go run .

# 指定主机和端口
go run . 0.0.0.0 6380
```

### 使用 redis-cli 连接测试

```bash
# 连接到默认端口
redis-cli -h 127.0.0.1 -p 6379

# 连接到指定端口
redis-cli -h 127.0.0.1 -p 6380
```

## 支持的命令

- `PING` - 返回 PONG
- `ECHO <message>` - 回显消息
- `SET <key> <value>` - 设置键值对
- `GET <key>` - 获取键对应的值
- `INFO` - 返回服务器信息
- `QUIT` - 断开连接

## 项目结构

```
goRedis/
├── main.go          # 主程序入口
├── server.go        # 服务器实现
├── resp.go          # RESP 协议实现
├── go.mod           # 模块文件
└── README.md        # 项目说明
```

## 协议支持

当前实现支持完整的 RESP 协议：

- **简单字符串** (`+`) - 状态消息
- **错误** (`-`) - 错误响应
- **整数** (`:`) - 数字值
- **批量字符串** (`$`) - 数据内容
- **数组** (`*`) - 命令和参数

### 命令格式示例

服务器支持标准的 Redis 命令格式：

```bash
# 使用 redis-cli（自动转换为 RESP 数组格式）
redis-cli> PING
PONG

redis-cli> SET mykey myvalue
OK

redis-cli> GET mykey
"myvalue"

redis-cli> ECHO "hello world"
"hello world"
```

## 下一步计划

- 添加更多 Redis 命令 (DEL, EXISTS, KEYS 等)
- 实现过期时间 (TTL, EXPIRE)
- 支持更多数据类型 (List, Hash, Set)
- 实现持久化 (RDB, AOF)
- 添加配置文件和日志系统