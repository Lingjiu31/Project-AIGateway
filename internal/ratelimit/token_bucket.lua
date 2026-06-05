-- KEYS[1]  - Redis key，格式为 ratelimit:{userID}
-- ARGV[1]  - rate：每秒补充的令牌数（浮点）
-- ARGV[2]  - burst：桶容量上限（浮点）
-- ARGV[3]  - now：当前时间戳（微秒）
--
-- 返回 1 表示放行，0 表示拒绝

local key       = KEYS[1]
local rate      = tonumber(ARGV[1])
local burst     = tonumber(ARGV[2])
local now       = tonumber(ARGV[3])

-- 读取当前状态，第一次不存在返回 false，用默认值
local last      = redis.call('HMGET', key, 'tokens', 'last_time')
local tokens    = tonumber(last[1]) or burst   -- 第一次默认满桶
local last_time = tonumber(last[2]) or now     -- 第一次默认当前时间

-- 距上次请求过了多少秒，然后按比例补充令牌
local elapsed   = (now - last_time) / 1e6      -- 微秒转秒
local new_tokens = math.min(burst, tokens + elapsed * rate)

-- 令牌不足，拒绝
if new_tokens < 1 then
    redis.call('HMSET', key, 'tokens', new_tokens, 'last_time', now)
    redis.call('EXPIRE', key, 3600)
    return 0
end

-- 消耗一个令牌，放行
redis.call('HMSET', key, 'tokens', new_tokens - 1, 'last_time', now)
redis.call('EXPIRE', key, 3600)
return 1