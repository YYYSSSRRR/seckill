package service

import (
	"context"
	v1 "gateway/api/gateway/v1"

	"gateway/internal/biz"
	v1pro1 "proto_definitions/product/v1"
	v1user1 "proto_definitions/user/v1"
)

type GatewayService struct {
	v1.UnimplementedGatewayServiceServer
	GatewayUsecase *biz.GatewayUsecase
}

func NewGatewayService(gu *biz.GatewayUsecase) *GatewayService {
	return &GatewayService{GatewayUsecase: gu}
}

func (gs *GatewayService) GetProductInfo(ctx context.Context, request *v1pro1.QueryRequest) (*v1pro1.ProductInfoResponse, error) {
	queryReq := &v1pro1.QueryRequest{Id: request.Id}
	resp, err := gs.GatewayUsecase.Product.GetProductInfo(ctx, queryReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (gs *GatewayService) Login(ctx context.Context, in *v1user1.LoginRequest) (*v1user1.LoginResponse, error) {
	req := &v1user1.LoginRequest{Email: in.Email, Password: in.Password}
	resp, err := gs.GatewayUsecase.User.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
