package service

import (
	"context"
	"errors"
	"fmt"
	"product_service/internal/biz"
	v1 "proto_definitions/product/v1"

	"github.com/shopspring/decimal"
)

type ProductService struct {
	v1.UnimplementedProductServiceServer
	productUsercase *biz.ProductUsecase
}

func NewProductService(pu *biz.ProductUsecase) *ProductService {
	return &ProductService{productUsercase: pu}
}

func (ps *ProductService) AddProduct(ctx context.Context, req *v1.AddProductRequest) (*v1.ProductResponse, error) {
	name := req.Name
	description := req.Describe
	priceCents := req.Price
	stock := req.Stock
	price := decimal.NewFromInt(priceCents).Div(decimal.NewFromInt(int64(100)))

	product := &biz.Product{Name: name, Describe: description, Price: price, Stock: int(stock)}

	if !product.IsPriceValid() || !product.IsStockValid() {
		return &v1.ProductResponse{Success: false}, errors.New("参数不合法，请重试")
	}

	err := ps.productUsercase.AddProduct(ctx, product)
	if err != nil {
		return &v1.ProductResponse{Success: false}, err
	}

	return &v1.ProductResponse{Success: true}, nil
}

func (ps *ProductService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.ProductResponse, error) {
	id := req.Id
	num := req.Num
	ok, err := ps.productUsercase.DeductStock(ctx, id, int(num))
	if err != nil {
		fmt.Printf("扣库存失败")
		return &v1.ProductResponse{Success: false}, err
	}
	if !ok {
		fmt.Printf("扣库存失败，ok=false")
		return &v1.ProductResponse{Success: false}, err
	}
	fmt.Printf("扣库存：%d", req.Num)
	return &v1.ProductResponse{Success: ok}, nil
}

func (ps *ProductService) AddStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.ProductResponse, error) {
	id := req.Id
	num := req.Num
	err := ps.productUsercase.AddStock(ctx, id, int(num))
	if err != nil {
		return &v1.ProductResponse{Success: false}, err
	}
	fmt.Printf("加库存：%d", req.Num)
	return &v1.ProductResponse{Success: true}, nil
}

func (ps *ProductService) GetProductInfo(ctx context.Context, req *v1.QueryRequest) (*v1.ProductInfoResponse, error) {
	id := req.Id
	product, err := ps.productUsercase.GetProductInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	priceCents := product.Price.IntPart() * 100

	return &v1.ProductInfoResponse{
		Id:       id,
		Name:     product.Name,
		Describe: product.Describe,
		Price:    priceCents,
		Stock:    int32(product.Stock),
	}, nil

}

func (ps *ProductService) EditProductPrice(ctx context.Context, req *v1.EditRequest) (*v1.ProductResponse, error) {
	id := req.Id
	priceCents := req.Price
	price := decimal.NewFromInt(priceCents).Div(decimal.NewFromInt(int64(100)))
	err := ps.productUsercase.EditProductPrice(ctx, id, price)
	if err != nil {
		return &v1.ProductResponse{Success: false}, err
	}

	return &v1.ProductResponse{Success: true}, nil
}
