-- KEYS[1]="order:lock:product:ProductID"
-- ARGV[1]=userID
-- 用集合存下单product的user
local result=redis.call('SADD',KEYS[1],ARGV[1])
if result==1 then
    -- 添加成功
    return 1
else
    -- 添加失败
    return 0
end

