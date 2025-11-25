package data

import (
	"context"
	"log"
	v1 "proto_definitions/product/v1"
	"seckill_service/internal/biz"
	"seckill_service/internal/conf"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type productRepo struct {
	client v1.ProductServiceClient
}

func NewProductRepo(pc v1.ProductServiceClient) biz.ProductRepo {

	return &productRepo{client: pc}
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
	client := v1.NewProductServiceClient(conn)

	return client, nil
}

func (pr *productRepo) DeductStock(ctx context.Context, in *v1.DeductStockRequest) (*v1.ProductResponse, error) {
	resp, err := pr.client.DeductStock(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (pr *productRepo) GetProductInfo(ctx context.Context, in *v1.QueryRequest) (*v1.ProductInfoResponse, error) {
	log.Printf("Calling Product service for product ID: %d", in.Id)
	resp, err := pr.client.GetProductInfo(ctx, in)
	if err != nil {
		log.Printf("Product service call failed: %v", err)
		return nil, status.Errorf(codes.Aborted, "wait result not return success: FAILURE")
	}

	return resp, nil
}

func (pr *productRepo) AddStock(ctx context.Context, in *v1.DeductStockRequest) (*v1.ProductResponse, error) {
	resp, err := pr.client.AddStock(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
