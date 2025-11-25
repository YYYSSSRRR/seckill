package biz

import (
	"context"
	"errors"
	"user_service/internal/utils"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// User 就是领域对象，承载了核心业务数据和可拓展的领域规则，且独立于外部依赖
type User struct {
	ID       int64
	Email    string
	Password string
	Money    decimal.Decimal `gorm:"type:decimal(12,2);default:0.00"`
}

func (user *User) IsEmailValid() bool {
	return user.Email != "" && len(user.Email) > 0
}

// UserRepo 仓储接口，直接调用其中的方法访问数据库操作
type UserRepo interface {
	//密码校验在应用层检查，data层只进行查询数据库，不负责逻辑判断
	GetUserByEmail(email string) (*User, error)

	CreateUser(user *User) error

	//可以充值或扣款，在sql条件判断时加上money>0，返回值为RowsAffected就可以判断是否扣款成功
	UpdateMoney(id int64, price decimal.Decimal) int64

	//查询余额
	GetUserMoney(id int64) (decimal.Decimal, error)
}

// UserUsecase 业务用例接口（不是领域对象！），是业务逻辑的编排，使用领域对象和仓储接口完成业务需求，实现业务逻辑
type UserUsecase interface {
	//找到了email就是注册,否则就是登录，成功就返回token
	Login(ctx context.Context, email string, password string) (string, error)

	//用户充值
	ReCharge(ctx context.Context, id int64, money decimal.Decimal) error

	//用户消费
	Cost(ctx context.Context, id int64, money decimal.Decimal) error

	//查询余额
	GetBalance(ctx context.Context, id int64) (decimal.Decimal, error)
}

type userUsecase struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserUsecase(repo UserRepo, logger log.Logger) UserUsecase {
	return &userUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (u *userUsecase) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(email)
	if user != nil {
		correct := utils.CheckPassword(password, user.Password)
		if correct {
			return "token", nil
		}
		return "", nil
	}
	if err == nil {
		hashedPassword, err := utils.EncodePassword(password)
		if err != nil {
			return "", err
		}
		user := &User{Email: email, Password: hashedPassword}

		//这里就是利用领域对象自己判断参数是否合法
		if user.IsEmailValid() {
			err := u.repo.CreateUser(user)
			if err != nil {
				return "", err
			}
			return "token", nil
		}
		return "", errors.New("参数不合法，请重新输入邮箱和密码")

	}

	return "", err

}

func (u *userUsecase) ReCharge(ctx context.Context, id int64, money decimal.Decimal) error {
	result := u.repo.UpdateMoney(id, money)
	if result > 0 {
		return nil
	}
	return errors.New("余额不足")
}

func (u *userUsecase) Cost(ctx context.Context, id int64, money decimal.Decimal) error {
	result := u.repo.UpdateMoney(id, money.Neg())
	if result > 0 {
		return nil
	}
	return errors.New("余额不足")
}

func (u *userUsecase) GetBalance(ctx context.Context, id int64) (decimal.Decimal, error) {
	money, err := u.repo.GetUserMoney(id)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return money, nil
}
