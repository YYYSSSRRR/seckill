package service

import (
	"context"
	"errors"
	"fmt"
	v1 "proto_definitions/user/v1"
	"user_service/internal/biz"

	"github.com/shopspring/decimal"
)

// UserService 要实现接口proto定义的方法，所以不需要写接口
type UserService struct {
	v1.UnimplementedUserServiceServer
	uc biz.UserUsecase
}

func NewUserService(uc biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (us *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	token, err := us.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	fmt.Printf("登录中")

	return &v1.LoginResponse{Token: token}, nil
}

func (us *UserService) GetUserById(ctx context.Context, req *v1.UserInfoRequest) (*v1.UserInfoResponse, error) {
	balance, err := us.uc.GetBalance(ctx, req.Id)

	if err != nil {
		return nil, err
	}

	balanceCents := balance.Mul(decimal.NewFromInt(100))
	if !balanceCents.IsInteger() {
		return nil, errors.New("余额格式错误，存在分一下的小数")
	}
	money := balanceCents.IntPart()
	return &v1.UserInfoResponse{Money: money}, nil

}

func (us *UserService) CostMoney(ctx context.Context, req *v1.UserInfoRequest) (*v1.UserChargeResponse, error) {
	//存储时是用分来存储，要转换成元（decimal类型）
	money := decimal.NewFromInt(req.Money).Div(decimal.NewFromInt(100))
	err := us.uc.Cost(ctx, req.Id, money)
	if err != nil {
		return &v1.UserChargeResponse{Success: false}, nil
	}
	fmt.Printf("扣钱：%d", money)
	return &v1.UserChargeResponse{Success: true}, nil
}

func (us *UserService) RechargeMoney(ctx context.Context, req *v1.UserInfoRequest) (*v1.UserChargeResponse, error) {
	//存储时是用分来存储，要转换成元（decimal类型）
	money := decimal.NewFromInt(req.Money).Div(decimal.NewFromInt(100))
	err := us.uc.ReCharge(ctx, req.Id, money)
	if err != nil {
		return &v1.UserChargeResponse{Success: false}, nil
	}
	fmt.Printf("加钱：%d", money)
	return &v1.UserChargeResponse{Success: true}, nil
}
