package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	v1 "proto_definitions/product/v1"
	"time"

	//clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:9001", "gRPC server address")
)

func main() {
	//cli, err := clientv3.New(clientv3.Config{
	//	Endpoints:   []string{"127.0.0.1:2379"},
	//	DialTimeout: 5 * time.Second,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//defer cli.Close()
	//
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//defer cancel()
	//
	//// Kratos 默认前缀是 /microservices/
	//resp, err := cli.Get(ctx, "/microservices/", clientv3.WithPrefix())
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("📦 已注册服务列表:")
	//for _, kv := range resp.Kvs {
	//	fmt.Printf("%s = %s\n", kv.Key, kv.Value)
	//}

	// 1. 连接 gRPC 服务（无 TLS 加密，生产环境需配置 TLS）
	conn, err := grpc.Dial(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // 等待连接成功
	)
	if err != nil {
		log.Fatalf("无法连接服务：%v", err)
	}
	defer conn.Close()

	// 2. 创建客户端实例
	client := v1.NewProductServiceClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 3. 根据命令执行对应操作
	//addProduct(ctx, client)
	getProductInfo(ctx, client)
	//deductStock(ctx, client)
	//editProductPrice(ctx, client)
}

// 新增商品
func addProduct(ctx context.Context, client v1.ProductServiceClient) {
	req := &v1.AddProductRequest{
		Name:     "测试商品",
		Describe: "这是一个用于 gRPC 客户端测试的商品",
		Price:    99900,
		Stock:    100,
	}

	resp, err := client.AddProduct(ctx, req)
	if err != nil {
		log.Fatalf("新增商品失败：%v", err)
	}
	fmt.Printf("新增商品结果：%v\n", resp.Success)
}

// 扣减库存
func deductStock(ctx context.Context, client v1.ProductServiceClient) {
	req := &v1.DeductStockRequest{
		Id:  1,  // 替换为实际商品 ID
		Num: 10, // 扣减数量
	}

	resp, err := client.DeductStock(ctx, req)
	if err != nil {
		log.Fatalf("扣减库存失败：%v", err)
	}
	fmt.Printf("扣减库存结果：%v\n", resp.Success)
}

// 增加库存
func addStock(ctx context.Context, client v1.ProductServiceClient) {
	req := &v1.DeductStockRequest{ // 复用 DeductStockRequest 结构体（字段一致）
		Id:  1,  // 替换为实际商品 ID
		Num: 20, // 增加数量
	}

	resp, err := client.AddStock(ctx, req)
	if err != nil {
		log.Fatalf("增加库存失败：%v", err)
	}
	fmt.Printf("增加库存结果：%v\n", resp.Success)
}

// 查询商品信息
func getProductInfo(ctx context.Context, client v1.ProductServiceClient) {
	req := &v1.QueryRequest{
		Id: 1,
	}

	resp, err := client.GetProductInfo(ctx, req)
	if err != nil {
		log.Fatalf("查询商品失败：%v", err)
	}
	fmt.Printf("商品信息：\n")
	fmt.Printf("ID: %d\n", resp.Id)
	fmt.Printf("名称: %s\n", resp.Name)
	fmt.Printf("描述: %s\n", resp.Describe)
	fmt.Printf("价格: %d\n", resp.Price)
	fmt.Printf("库存: %d\n", resp.Stock)
}

// 修改商品价格
func editProductPrice(ctx context.Context, client v1.ProductServiceClient) {
	req := &v1.EditRequest{
		Id:    1,   // 替换为实际商品 ID
		Price: 888, // 新价格
	}

	resp, err := client.EditProductPrice(ctx, req)
	if err != nil {
		log.Fatalf("修改价格失败：%v", err)
	}
	fmt.Printf("修改价格结果：%v\n", resp.Success)

}
