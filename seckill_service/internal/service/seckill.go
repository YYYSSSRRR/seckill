package service

import (
	"context"
	v1 "proto_definitions/seckill/v1"
	"seckill_service/internal/biz"
)

type SeckillService struct {
	su biz.SeckillUsecase
	v1.UnimplementedSeckillServiceServer
}

func NewSeckillService(su biz.SeckillUsecase) *SeckillService {
	return &SeckillService{su: su}
}

func (ss *SeckillService) Seckill(ctx context.Context, in *v1.SeckillRequest) (*v1.SeckillResponse, error) {

	resp, err := ss.su.Seckill(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
