package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"product_service/internal/biz"
	"product_service/internal/utils"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type productRepo struct {
	data        *Data
	redisClient *RedisClient
}

func NewProductRepo(data *Data, redisClient *RedisClient) biz.ProductRepo {
	return &productRepo{data: data, redisClient: redisClient}
}

func (p *productRepo) AddProduct(product *biz.Product) error {
	result := p.data.gormDB.Create(product)
	if result.Error != nil {
		return result.Error
	}

	//热点key先加到缓存中
	cacheKey := utils.PRODUCT_CACHE + strconv.FormatInt(product.ID, 10)
	productJson, err := json.Marshal(product)
	if err != nil {
		return err
	}
	//先设置普通过期
	p.redisClient.client.Set(context.Background(), cacheKey, string(productJson), time.Hour*3)

	return nil
}

func (p *productRepo) GetProductInfo(id int64) (*biz.Product, error) {
	ctx := context.Background()
	//查数据库时需要加分布式锁，因为热点key过期时会有大量数据打到数据库
	product := &biz.Product{}

	//先查redis
	result := p.redisClient.client.Get(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10))

	val, err := result.Result()

	//查询到了数据，直接返回
	if val != "" {
		if val == "Not Found" {
			return nil, errors.New("没查询到数据")
		}
		err := json.Unmarshal([]byte(val), product)
		if err != nil {
			return nil, err
		}
		return product, nil
	}

	if err != redis.Nil {
		return nil, err
	}
	//没查询到缓存，要加锁查询数据库

	//分布式锁，先加key到redis中，且每一把锁都要有不同的uuid
	lockUUID := uuid.New()
	ok := p.redisClient.client.SetNX(ctx, utils.LOCK_KEY+strconv.FormatInt(id, 10), lockUUID, time.Second*10)

	//如果获取锁成功，先判断缓存中是否有数据
	if ok.Val() {
		result, err := p.redisClient.client.Get(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10)).Result()

		if err != nil && err != redis.Nil {
			return nil, err
		}

		//查询到了缓存中的数据：直接返回
		if result != "" {
			if result == "Not Found" {
				return nil, errors.New("没查询到数据")
			}
			err := json.Unmarshal([]byte(val), product)
			if err != nil {
				return nil, err
			}
			return product, nil
		}

		//没查询到缓存中的数据：查询数据库
		result1 := p.data.gormDB.First(product, "id=?", id)
		if result1.Error == gorm.ErrRecordNotFound {
			//数据库中也没有，存空缓存，防止缓存穿透
			p.redisClient.client.Set(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10), "Not Found", time.Second*3)
			return nil, errors.New("没查询到数据")
		}

		//查询到数据了：存到数据库中
		resultJson, err := json.Marshal(product)
		if err != nil {
			return nil, err
		}
		p.redisClient.client.Set(ctx, utils.PRODUCT_CACHE+strconv.FormatInt(id, 10), string(resultJson), time.Hour*3)

		//查询完了，释放锁(先判断是不是自己的锁，然后再释放)
		scriptContent, err := utils.LoadLuaScript("/Users/ysr/Documents/seckill_microservice/product_service/internal/lua/unlock.lua")
		if err != nil {
			return nil, err
		}
		lockValue, err := p.redisClient.client.Get(ctx, utils.LOCK_KEY+strconv.FormatInt(id, 10)).Result()
		if err != nil {
			return nil, err
		}
		result2, err := p.redisClient.client.Eval(ctx, scriptContent, []string{lockValue}, lockUUID.String()).Result()
		if err != nil {
			return nil, err
		}
		if result2 == 1 {
			fmt.Printf("释放锁成功")
		}
		if result2 == 0 {
			fmt.Printf("释放锁失败")
		}

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
	result := p.data.gormDB.Exec(`update products set stock=stock+? where id=? and stock-?>=0`, num, id, num)
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
