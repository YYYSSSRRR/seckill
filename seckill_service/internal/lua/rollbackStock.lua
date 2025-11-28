-- 把扣减的库存加回来
-- 下单失败，把用户的id从集合中删除
-- KEYS[1]:商品的key
-- KEYS[2]:存商品秒杀用户列表的key
-- ARGV[1]:用户id
local stock=redis.call('hget',KEYS[1],"stock")
redis.call('HSET',KEYS[1],"stock",stock+1)
redis.call('SREM',KEYS[2],ARGV[1])
return 1