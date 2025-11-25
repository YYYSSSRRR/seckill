package data

import (
	"user_service/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// userRepo 仓储接口的实现：用于访问数据库的操作，领域层可以直接调用UserRepo中的方法访问数据库
type userRepo struct {
	Data *Data
}

// 实现biz层的UserRepo接口，进行数据库的操作
func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{Data: data}
}

func (ur *userRepo) GetUserByEmail(email string) (*biz.User, error) {
	user := &biz.User{}
	result := ur.Data.gormDB.Where("email=?", email).First(user)
	//记录没找到
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	//其他错误，直接返回
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (ur *userRepo) CreateUser(user *biz.User) error {

	result := ur.Data.gormDB.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func (ur *userRepo) UpdateMoney(id int64, price decimal.Decimal) int64 {
	//需要在sql中判断是否余额大于0,不能单独判断不然不能保证原子性
	result := ur.Data.gormDB.Table("users").Exec(`update users set money=money+? where id=? and (money+?)>=0`, price, id, price)

	//如果返回值>0就是插入成功
	return result.RowsAffected
}

func (ur *userRepo) GetUserMoney(id int64) (decimal.Decimal, error) {
	money := decimal.Decimal{}
	row := ur.Data.gormDB.Raw("select money from users where id=?", id).Row()
	err := row.Scan(&money)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return money, nil
}
