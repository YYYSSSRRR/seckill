package biz

import (
	"context"

	"github.com/shopspring/decimal"
)

type Product struct {
	ID       int64
	Name     string
	Describe string
	Price    decimal.Decimal `gorm:"type:decimal(12,2)"`
	Stock    int
}

// Product是领域对象
func (product *Product) IsPriceValid() bool {
	return product.Price.GreaterThan(decimal.NewFromInt(int64(0))) && product.Price.IsInteger()
}

func (product *Product) IsStockValid() bool {
	return product.Stock >= 0
}

type ProductRepo interface {
	//新增商品
	AddProduct(product *Product) error

	//查商品信息
	//TODO 添加热点key缓存
	GetProductInfo(id int64) (*Product, error)

	//减商品库存
	//TODO 防止超卖问题
	DeductStock(id int64, num int) (bool, error)

	//增加商品库存
	AddStock(id int64, num int) error

	//修改商品价格
	EditProductPrice(id int64, price decimal.Decimal) error
}

type ProductUsecase struct {
	pr ProductRepo
}

func NewProductUsecase(pr ProductRepo) *ProductUsecase {
	return &ProductUsecase{pr: pr}
}

func (pu *ProductUsecase) AddProduct(ctx context.Context, product *Product) error {
	err := pu.pr.AddProduct(product)
	if err != nil {
		return err
	}
	return nil
}

func (pu *ProductUsecase) GetProductInfo(ctx context.Context, id int64) (*Product, error) {
	product, err := pu.pr.GetProductInfo(id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (pu *ProductUsecase) DeductStock(ctx context.Context, id int64, num int) (bool, error) {
	success, err := pu.pr.DeductStock(id, num)
	if err != nil {
		return false, err
	}
	return success, nil
}

func (pu *ProductUsecase) AddStock(ctx context.Context, id int64, num int) error {
	err := pu.pr.AddStock(id, num)
	if err != nil {
		return err
	}
	return nil
}

func (pu *ProductUsecase) EditProductPrice(ctx context.Context, id int64, price decimal.Decimal) error {
	err := pu.pr.EditProductPrice(id, price)
	if err != nil {
		return err
	}
	return nil
}
