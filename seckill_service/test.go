package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "proto_definitions/seckill/v1"
)

func main() {
	// 通过命令行参数指定服务端地址
	serverAddr := flag.String("server_addr", "localhost:9000", "The server address in the format of host:port")
	flag.Parse()

	fmt.Printf("Connecting to gRPC server at: %s\n", *serverAddr)

	// 1. 设置 gRPC 连接
	conn, err := grpc.Dial(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // 等待连接建立
	)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to gRPC server")

	// 2. 创建 gRPC 客户端实例
	client := pb.NewSeckillServiceClient(conn)

	// 3. 准备请求数据
	request := &pb.SeckillRequest{
		UserID:    1, // 示例用户 ID
		ProductID: 1, // 示例商品 ID
	}

	// 4. 设置请求上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 延长超时时间
	defer cancel()

	// 5. 调用 gRPC 服务的 Seckill 方法
	fmt.Printf("Sending seckill request: UserID=%d, ProductID=%d\n", request.UserID, request.ProductID)

	start := time.Now()
	response, err := client.Seckill(ctx, request)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Request failed after %v", duration)
		log.Printf("Error details:")
		log.Printf("  Error: %v", err)

		// 解析 gRPC 状态错误
		if st, ok := status.FromError(err); ok {
			log.Printf("  gRPC Status Code: %v", st.Code())
			log.Printf("  gRPC Status Message: %v", st.Message())
			log.Printf("  gRPC Status Details: %v", st.Details())
		}

		log.Fatalf("Seckill call failed")
	}

	// 6. 处理并打印响应结果
	fmt.Printf("Request succeeded in %v\n", duration)
	fmt.Println("Seckill request successful!")
	fmt.Printf("Received response: OrderID=%d, UserID=%d, ProductID=%d, Price=%d\n",
		response.OrderID,
		response.UserID,
		response.ProductID,
		response.Price)
}
