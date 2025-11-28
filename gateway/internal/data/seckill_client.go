package data

import (
	"context"
	"gateway/internal/biz"
	v1 "proto_definitions/seckill/v1"
)

type seckillRepo struct {
	client v1.SeckillServiceClient
}

func NewSeckillRepo(client v1.SeckillServiceClient) biz.SeckillRepo {
	return &seckillRepo{client: client}
}

func (sr *seckillRepo) Seckill(ctx context.Context, in *v1.SeckillRequest) (*v1.SeckillResponse, error) {
	return sr.client.Seckill(ctx, in)
}

func NewSeckillClient()
