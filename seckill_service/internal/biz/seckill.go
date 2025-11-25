package biz

import (
	"context"
	"fmt"
	v1pro "proto_definitions/product/v1"
	v1sec "proto_definitions/seckill/v1"
	v1user "proto_definitions/user/v1"
	"seckill_service/internal/conf"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v4"
	"github.com/shopspring/decimal"
)

// Order 领域模型
type Order struct {
	OrderID   int64 `gorm:"column:id;primaryKey"`
	ProductID int64
	UserID    int64
	Price     decimal.Decimal
	PayStatus int8
}

type SeckillRepo interface {
	CreateOrder(order *Order) (bool, error)
	PayOrder(orderID int64) (bool, error)
	QueryOrder(orderID int64) (*Order, error)
}

// 定义仓储接口，在data层实现方法（rpc调用user_service的服务）
type UserRepo interface {
	//用户消费
	CostMoney(ctx context.Context, in *v1user.UserInfoRequest) (*v1user.UserChargeResponse, error)
	RechargeMoney(ctx context.Context, in *v1user.UserInfoRequest) (*v1user.UserChargeResponse, error)
}

type ProductRepo interface {
	//扣减商品库存
	DeductStock(ctx context.Context, in *v1pro.DeductStockRequest) (*v1pro.ProductResponse, error)
	AddStock(ctx context.Context, in *v1pro.DeductStockRequest) (*v1pro.ProductResponse, error)

	//返回商品信息
	GetProductInfo(ctx context.Context, in *v1pro.QueryRequest) (*v1pro.ProductInfoResponse, error)
}

type SeckillUsecase interface {
	Seckill(ctx context.Context, in *v1sec.SeckillRequest) (*v1sec.SeckillResponse, error)
}

type seckillUsecase struct {
	User    UserRepo
	Product ProductRepo
	Sec     SeckillRepo
	conf    *conf.Registry
}

func NewSeckillUsecase(user UserRepo, product ProductRepo, seckill SeckillRepo, conf *conf.Registry) SeckillUsecase {
	return &seckillUsecase{User: user, Product: product, Sec: seckill, conf: conf}
}

func (su *seckillUsecase) Seckill(ctx context.Context, in *v1sec.SeckillRequest) (*v1sec.SeckillResponse, error) {
	productID := in.ProductID
	userID := in.UserID

	//先查询商品价格
	queryReq := &v1pro.QueryRequest{Id: productID}
	productInfo, err := su.Product.GetProductInfo(ctx, queryReq)
	if err != nil {
		return nil, err
	}
	fmt.Printf("商品信息：%v", productInfo.Price)
	priceCents := productInfo.Price

	//开启saga事务，扣库存、减客户的钱，阻塞等待事务成功，最后创建订单
	dtmServer := "localhost:36790"
	saga := dtmgrpc.NewSagaGrpc(dtmServer, shortuuid.New()).
		//扣商品库存
		Add("10.130.97.157:9001/product.v1.ProductService/DeductStock",
			"10.130.97.157:9001/product.v1.ProductService/AddStock",
			&v1pro.DeductStockRequest{Id: productID, Num: 1},
		).
		//减客户的钱
		Add("10.130.97.157:9002/user.v1.UserService/CostMoney",
			"10.130.97.157:9002/user.v1.UserService/RechargeMoney",
			&v1user.UserInfoRequest{Id: userID, Money: priceCents},
		)
	saga.WaitResult = true
	err3 := saga.Submit()
	if err3 != nil {
		return nil, err3
	}

	price := decimal.NewFromInt(priceCents).Div(decimal.NewFromInt(int64(100)))
	order := &Order{ProductID: productID, UserID: userID, Price: price}
	seckillRepo := su.Sec
	_, err4 := seckillRepo.CreateOrder(order)
	if err4 != nil {
		return nil, err4
	}

	seckillResponse := &v1sec.SeckillResponse{ProductID: productID, UserID: userID, Price: priceCents, OrderID: order.OrderID}

	return seckillResponse, nil
}
