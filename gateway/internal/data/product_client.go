package data

import (
	"context"
	"gateway/internal/biz"
	"gateway/internal/conf"
	v1 "proto_definitions/product/v1"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type productRepo struct {
	client v1.ProductServiceClient
}

func NewProductRepo(client v1.ProductServiceClient) biz.ProductRepo {
	return &productRepo{client: client}
}

func NewProductServiceClient(r registry.Discovery, conf *conf.Registry) (v1.ProductServiceClient, error) {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(conf.Etcd.Discovery.Product),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		return nil, err
	}
	return v1.NewProductServiceClient(conn), nil
}

func (pr *productRepo) GetProductInfo(ctx context.Context, in *v1.QueryRequest) (*v1.ProductInfoResponse, error) {
	resp, err := pr.client.GetProductInfo(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
