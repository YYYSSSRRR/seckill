package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	//// 从命令行参数获取 gRPC 服务地址（默认本地 9000 端口）
	//addr := flag.String("addr", "127.0.0.1:9000", "gRPC 服务地址")
	//flag.Parse()
	//
	//// 1. 建立 gRPC 连接（非安全连接，适合测试）
	//conn, err := grpc.Dial(*addr,
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	grpc.WithBlock(), // 等待连接成功
	//	grpc.WithTimeout(5*time.Second),
	//)
	//if err != nil {
	//	log.Fatalf("无法连接到 gRPC 服务: %v", err)
	//}
	//defer conn.Close()
	//
	//// 2. 创建 UserService 客户端
	//client := v1.NewUserServiceClient(conn)
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//
	//// 3. 测试 Login 接口（登录/注册，获取 token）
	//fmt.Println("=== 测试 Login 接口 ===")
	//loginReq := &v1.LoginRequest{
	//	Email:    "test@example.com", // 测试邮箱
	//	Password: "123456",           // 测试密码（实际应加密，此处为示例）
	//}
	//loginResp, err := client.Login(ctx, loginReq)
	//if err != nil {
	//	log.Fatalf("Login 失败: %v", err)
	//}
	//fmt.Printf("Login 成功，token: %s\n\n", loginResp.Token)
	//
	//// 4. 测试 GetUserById 接口（查询用户余额）
	//fmt.Println("=== 测试 GetUserById 接口 ===")
	//// 假设登录后获取到用户 ID 为 1001（实际应从 token 解析，此处简化）
	//userID := int64(1)
	//getReq := &v1.UserInfoRequest{Id: userID}
	//getResp, err := client.GetUserById(ctx, getReq)
	//if err != nil {
	//	log.Fatalf("查询用户余额失败: %v", err)
	//}
	//fmt.Printf("用户 %d 当前余额: %d 分\n\n", userID, getResp.Money)
	//
	//// 5. 测试 RechargeMoney 接口（充值）
	//fmt.Println("=== 测试 RechargeMoney 接口 ===")
	//rechargeAmount := int64(1000) // 充值 1000 分（即 10 元）
	//rechargeReq := &v1.UserInfoRequest{
	//	Id:    userID,
	//	Money: rechargeAmount,
	//}
	//rechargeResp, err := client.RechargeMoney(ctx, rechargeReq)
	//if err != nil {
	//	log.Fatalf("充值失败: %v", err)
	//}
	//if rechargeResp.Success {
	//	fmt.Printf("充值成功，已充值 %d 分\n", rechargeAmount)
	//} else {
	//	fmt.Println("充值失败")
	//}
	//
	//// 充值后再次查询余额，验证是否生效
	//getRespAfterRecharge, _ := client.GetUserById(ctx, getReq)
	//fmt.Printf("充值后余额: %d 分\n\n", getRespAfterRecharge.Money)
	//
	//// 6. 测试 CostMoney 接口（扣钱，模拟购买商品）
	//fmt.Println("=== 测试 CostMoney 接口 ===")
	//costAmount := int64(500) // 扣钱 500 分（即 5 元）
	//costReq := &v1.UserInfoRequest{
	//	Id:    userID,
	//	Money: costAmount,
	//}
	//costResp, err := client.CostMoney(ctx, costReq)
	//if err != nil {
	//	log.Fatalf("扣钱失败: %v", err)
	//}
	//if costResp.Success {
	//	fmt.Printf("扣钱成功，已扣除 %d 分\n", costAmount)
	//} else {
	//	fmt.Println("扣钱失败（可能余额不足）")
	//}
	//
	//// 扣钱后再次查询余额，验证是否生效
	//getRespAfterCost, _ := client.GetUserById(ctx, getReq)
	//fmt.Printf("扣钱后余额: %d 分\n", getRespAfterCost.Money)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Kratos 默认前缀是 /microservices/
	resp, err := cli.Get(ctx, "/microservices/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	fmt.Println("📦 已注册服务列表:")
	for _, kv := range resp.Kvs {
		fmt.Printf("%s = %s\n", kv.Key, kv.Value)
	}
}
