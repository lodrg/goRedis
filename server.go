package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// RedisServer 表示 Redis 服务器
type RedisServer struct {
	host  string
	port  int
	store map[string]string
	mutex sync.RWMutex
}

// NewRedisServer 创建新的 Redis 服务器实例
func NewRedisServer(host string, port int) *RedisServer {
	return &RedisServer{
		host:  host,
		port:  port,
		store: make(map[string]string),
	}
}

// Start 启动服务器
func (rs *RedisServer) Start() error {
	address := fmt.Sprintf("%s:%d", rs.host, rs.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Redis server listening on %s\n", address)
	fmt.Println("Press Ctrl+C to stop the server")

	// 接受客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// 为每个连接启动一个 goroutine 处理
		go rs.handleConnection(conn)
	}
}

// handleConnection 处理客户端连接
func (rs *RedisServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	reader := bufio.NewReader(conn)

	for {
		// 解析 RESP 命令
		command, err := ParseRESP(reader)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client disconnected: %s\n", clientAddr)
				break
			}
			log.Printf("Error parsing command from %s: %v\n", clientAddr, err)
			// 发送错误响应
			errorResp := NewRESPValue(RESP_ERROR)
			errorResp.Str = "ERR " + err.Error()
			conn.Write(errorResp.SerializeRESP())
			continue
		}

		fmt.Printf("Received from %s: %s\n", clientAddr, command.ToString())

		// 处理命令
		response := rs.processCommand(command)
		conn.Write(response.SerializeRESP())
	}
}

// processCommand 处理 Redis 命令
func (rs *RedisServer) processCommand(command *RESPValue) *RESPValue {
	// 检查命令是否为数组类型
	if command.Type != RESP_ARRAY || command.IsNull {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR Protocol error: expected array"
		return errorResp
	}

	if len(command.Array) == 0 {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR empty command"
		return errorResp
	}

	// 获取命令名称
	cmdValue := command.Array[0]
	if cmdValue.Type != RESP_BULK_STRING {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR Protocol error: expected bulk string for command"
		return errorResp
	}

	cmd := strings.ToUpper(cmdValue.Str)

	switch cmd {
	case "PING":
		return rs.handlePing()
	case "ECHO":
		return rs.handleEcho(command)
	case "SET":
		return rs.handleSet(command)
	case "GET":
		return rs.handleGet(command)
	case "QUIT":
		return rs.handleQuit()
	case "INFO":
		return rs.handleInfo()
	default:
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR unknown command '" + cmd + "'"
		return errorResp
	}
}

// handlePing 处理 PING 命令
func (rs *RedisServer) handlePing() *RESPValue {
	resp := NewRESPValue(RESP_SIMPLE_STRING)
	resp.Str = "PONG"
	return resp
}

// handleEcho 处理 ECHO 命令
func (rs *RedisServer) handleEcho(command *RESPValue) *RESPValue {
	if len(command.Array) < 2 {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR wrong number of arguments for 'echo' command"
		return errorResp
	}

	// 检查参数类型
	if command.Array[1].Type != RESP_BULK_STRING {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR Protocol error: expected bulk string for echo argument"
		return errorResp
	}

	resp := NewRESPValue(RESP_BULK_STRING)
	resp.Str = command.Array[1].Str
	return resp
}

// handleSet 处理 SET 命令
func (rs *RedisServer) handleSet(command *RESPValue) *RESPValue {
	if len(command.Array) < 3 {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR wrong number of arguments for 'set' command"
		return errorResp
	}

	// 检查参数类型
	if command.Array[1].Type != RESP_BULK_STRING || command.Array[2].Type != RESP_BULK_STRING {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR Protocol error: expected bulk string for key and value"
		return errorResp
	}

	key := command.Array[1].Str
	value := command.Array[2].Str

	// 线程安全地设置键值对
	rs.mutex.Lock()
	rs.store[key] = value
	rs.mutex.Unlock()

	resp := NewRESPValue(RESP_SIMPLE_STRING)
	resp.Str = "OK"
	return resp
}

// handleGet 处理 GET 命令
func (rs *RedisServer) handleGet(command *RESPValue) *RESPValue {
	if len(command.Array) < 2 {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR wrong number of arguments for 'get' command"
		return errorResp
	}

	// 检查参数类型
	if command.Array[1].Type != RESP_BULK_STRING {
		errorResp := NewRESPValue(RESP_ERROR)
		errorResp.Str = "ERR Protocol error: expected bulk string for key"
		return errorResp
	}

	key := command.Array[1].Str

	// 线程安全地获取值
	rs.mutex.RLock()
	value, exists := rs.store[key]
	rs.mutex.RUnlock()

	if !exists {
		// 返回 null bulk string
		resp := NewRESPValue(RESP_BULK_STRING)
		resp.IsNull = true
		return resp
	}

	resp := NewRESPValue(RESP_BULK_STRING)
	resp.Str = value
	return resp
}

// handleQuit 处理 QUIT 命令
func (rs *RedisServer) handleQuit() *RESPValue {
	resp := NewRESPValue(RESP_SIMPLE_STRING)
	resp.Str = "OK"
	return resp
}

// handleInfo 处理 INFO 命令
func (rs *RedisServer) handleInfo() *RESPValue {
	resp := NewRESPValue(RESP_BULK_STRING)
	resp.Str = "# Server\r\nredis_version:0.1.0\r\n"
	return resp
}
