-- KEYS[1]是商品的key，查找库存
-- KEYS[2]是商品秒杀用户集合的key，用来记录用户是否下过单
-- ARGV[1]是userID
local stock= tonumber(redis.call('HGET', KEYS[1], "stock"))
-- 如果还有库存就扣减
if stock>0 then
    redis.call('HSET',KEYS[1],"stock",stock-1)
    return 1
else
    --没有库存就把用户从商品下单集合中删除
    redis.call('SREM',KEYS[2],ARGV[1])
    return 0
end