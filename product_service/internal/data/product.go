package data

import (
	"common/lock"
	"context"
	"errors"
	"fmt"
	"product_service/internal/biz"
	"product_service/internal/utils"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type productRepo struct {
	data        *Data
	redisClient *RedisClient
	log         log.Logger
}

func NewProductRepo(data *Data, redisClient *RedisClient, log log.Logger) biz.ProductRepo {
	return &productRepo{data: data, redisClient: redisClient, log: log}
}

func (p *productRepo) AddProduct(product *biz.Product) error {
	result := p.data.gormDB.Create(product)
	if result.Error != nil {
		return result.Error
	}

	//热点key先加到缓存中，用hash结构存
	cacheKey := utils.PRODUCT_CACHE + strconv.FormatInt(product.ID, 10)

	ctx := context.Background()
	//先设置普通过期
	//p.redisClient.client.Set(context.Background(), cacheKey, string(productJson), time.Hour*3)
	res := p.redisClient.client.HSet(
		ctx,
		cacheKey,
		"id", product.ID,
		"name", product.Name,
		"description", product.Describe,
		"price", product.Price,
		"stock", product.Stock,
	)
	p.redisClient.client.Expire(ctx, cacheKey, 3*time.Hour)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (p *productRepo) GetProductInfo(id int64) (*biz.Product, error) {
	ctx := context.Background()
	//查数据库时需要加分布式锁，因为热点key过期时会有大量数据打到数据库
	product := &biz.Product{}

	//先查redis
	result := p.redisClient.client.HGetAll(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10))
	//val是一个map
	val, err := result.Result()
	if err != nil {
		return nil, err
	}

	//查询到了数据，直接返回
	if len(val) != 0 {
		product.ID, err = strconv.ParseInt(val["id"], 10, 64)
		product.Name = val["name"]
		product.Describe = val["description"]

		if err != nil {
			return nil, err
		}
		product.Price = decimal.RequireFromString(val["price"])
		stockNum, err := strconv.ParseInt(val["stock"], 10, 64)
		if err != nil {
			return nil, err
		}
		product.Stock = int(stockNum)

		return product, nil
	}

	//没查询到缓存，要加分布式锁查询数据库
	//分布式锁，先加key到redis中，值是uuid（确保每把锁的值不同用来区分不同的锁）
	lockUUID := uuid.New().String()
	lockKey := utils.LOCK_KEY + strconv.FormatInt(id, 10)
	ok, err := lock.Lock(ctx, lockKey, lockUUID, time.Second*10, p.redisClient.client)

	//如果获取锁成功，先判断缓存中是否有数据
	if ok {
		//获取锁就要释放锁！！！不能因为获取到缓存中的数据就直接return忘记释放锁
		defer func() {
			_, err = lock.Unlock(ctx, lockKey, lockUUID, p.redisClient.client)
			if err != nil {
				err := p.log.Log(log.LevelError, "释放锁错误")
				if err != nil {
					return
				}
			}
		}()
		result, err := p.redisClient.client.HGetAll(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10)).Result()

		if err != nil {
			return nil, err
		}

		//查询到了缓存中的数据：直接返回
		if len(result) != 0 {
			product.ID, err = strconv.ParseInt(val["id"], 10, 64)
			product.Name = val["name"]
			product.Describe = val["description"]
			fmt.Printf("商品价格：%v", val["price"])

			product.Price = decimal.RequireFromString(val["price"])
			stockNum, err := strconv.ParseInt(val["stock"], 10, 64)
			if err != nil {
				return nil, err
			}
			product.Stock = int(stockNum)

			return product, nil
		}

		//没查询到缓存中的数据：查询数据库
		result1 := p.data.gormDB.First(product, "id=?", id)
		fmt.Printf("查询数据库：%v", result1)
		if result1.Error == gorm.ErrRecordNotFound {
			//TODO数据库中也没有，存空缓存，防止缓存穿透
			//p.redisClient.client.Set(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10), "Not Found", time.Second*3)
			return nil, errors.New("没查询到数据")
		}

		//查询到数据了：存到数据库中
		cacheKey := utils.PRODUCT_CACHE + strconv.FormatInt(product.ID, 10)
		p.redisClient.client.HSet(
			context.Background(),
			cacheKey,
			"id", product.ID,
			"name", product.Name,
			"description", product.Describe,
			"price", product.Price.String(),
			"stock", product.Stock,
		)

	} else {
		time.Sleep(100 * time.Millisecond)
		return p.GetProductInfo(id)
	}
	return product, nil
}

func (p *productRepo) DeductStock(id int64, num int) (bool, error) {
	result := p.data.gormDB.Exec(`update products set stock=stock-? where id=? and stock-?>=0`, num, id, num)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, errors.New("没库存啦")
	}
	return true, nil
}

func (p *productRepo) AddStock(id int64, num int) error {
	result := p.data.gormDB.Exec(`update products set stock=stock+? where id=?`, num, id, num)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (p *productRepo) EditProductPrice(id int64, price decimal.Decimal) error {
	result := p.data.gormDB.Exec(`update products set price=? where id=?`, price, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
