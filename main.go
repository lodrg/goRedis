package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	// 默认配置
	host := "127.0.0.1"
	port := 6379

	// 从命令行参数读取配置
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	if len(os.Args) > 2 {
		if p, err := strconv.Atoi(os.Args[2]); err == nil {
			port = p
		}
	}

	server := NewRedisServer(host, port)

	fmt.Printf("Starting Redis server on %s:%d\n", host, port)
	fmt.Println("Usage: go run . [host] [port]")
	fmt.Println("Example: go run . 127.0.0.1 6379")

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
