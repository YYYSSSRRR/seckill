package biz

import (
	"context"
	v1pro "proto_definitions/product/v1"
	v1user "proto_definitions/user/v1"
)

type UserRepo interface {
	Login(ctx context.Context, in *v1user.LoginRequest) (*v1user.LoginResponse, error)
}

type ProductRepo interface {
	GetProductInfo(ctx context.Context, in *v1pro.QueryRequest) (*v1pro.ProductInfoResponse, error)
}

type GatewayUsecase struct {
	User    UserRepo
	Product ProductRepo
}

func NewGatewayUsecase(user UserRepo, product ProductRepo) *GatewayUsecase {
	return &GatewayUsecase{User: user, Product: product}
}
