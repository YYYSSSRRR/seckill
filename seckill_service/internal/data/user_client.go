package data

import (
	"context"
	v1 "proto_definitions/user/v1"
	"seckill_service/internal/biz"
	"seckill_service/internal/conf"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type userRepo struct {
	client v1.UserServiceClient
}

func NewUserRepo(uc v1.UserServiceClient) biz.UserRepo {
	return &userRepo{client: uc}
}

func NewUserServiceClient(conf *conf.Registry, r registry.Discovery) (v1.UserServiceClient, error) {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(conf.Etcd.Discovery.User),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		return nil, err
	}
	return v1.NewUserServiceClient(conn), nil
}

func (ur *userRepo) CostMoney(ctx context.Context, in *v1.UserInfoRequest) (*v1.UserChargeResponse, error) {
	resp, err := ur.client.CostMoney(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (ur *userRepo) RechargeMoney(ctx context.Context, in *v1.UserInfoRequest) (*v1.UserChargeResponse, error) {
	resp, err := ur.client.RechargeMoney(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
