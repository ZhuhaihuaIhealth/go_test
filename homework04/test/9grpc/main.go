package main

import (
	"flag"
	"log"
	"os"
)

// "fmt"
// "net/http"

// "github.com/gin-gonic/gin"
// "google.golang.org/protobuf/proto"

// "protobuf/pb"
func main() {
	mode := flag.String("mode", "server", "运行模式: server/client")
	addr := flag.String("addr", "50051", "服务器地址(server模式)或连接地址(client模式)")
	flag.Parse()

	switch *mode {
	case "server":
		// 启动服务
		// 创建一个gRPC服务器
		// 监听端口
		// 等待客户端连接
		// 启动服务
		log.Println("启动grpc服务")
		if err := startServer(*addr); err != nil {
			log.Fatalf("启动grpc服务失败: %v", err)
		}

	case "client":
		log.Println("启动grpc客户端")
		if *addr == "50051" {
			*addr = "localhost:50051"
		}
		runClientDemo(*addr)
	default:
		log.Fatalf("无效的运行模式: %s", *mode)
		os.Exit(1)
	}

}
