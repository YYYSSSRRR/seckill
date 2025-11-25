package data

import (
	"context"
	"gateway/internal/biz"
	"gateway/internal/conf"
	v1 "proto_definitions/user/v1"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// data层实现远程连接
type userRepo struct {
	client v1.UserServiceClient
}

func NewUserRepo(client v1.UserServiceClient) biz.UserRepo {
	return &userRepo{client: client}
}

// grpc连接
func NewUserClientService(r registry.Discovery, conf *conf.Registry) (v1.UserServiceClient, error) {
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

func (ur *userRepo) Login(ctx context.Context, in *v1.LoginRequest) (*v1.LoginResponse, error) {
	resp, err := ur.client.Login(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
