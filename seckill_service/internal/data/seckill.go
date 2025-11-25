package data

import "seckill_service/internal/biz"

type seckillRepo struct {
	data *Data
}

func NewSeckillRepo(data *Data) biz.SeckillRepo {
	return &seckillRepo{data: data}
}

func (sr *seckillRepo) CreateOrder(order *biz.Order) (bool, error) {
	data1 := sr.data
	db := data1.gormDB
	result := db.Create(order)

	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (sr *seckillRepo) PayOrder(orderID int64) (bool, error) {
	return true, nil
}

func (sr *seckillRepo) QueryOrder(orderID int64) (*biz.Order, error) {
	return nil, nil
}
