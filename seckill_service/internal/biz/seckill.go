package biz

import (
	"context"
	"fmt"
	v1pro "proto_definitions/product/v1"
	v1sec "proto_definitions/seckill/v1"
	v1user "proto_definitions/user/v1"
	"seckill_service/internal/conf"

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

// 秒杀：先查询用户是否有购买资格（一人多单），再扣减库存，如果扣减成功就创建订单
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
