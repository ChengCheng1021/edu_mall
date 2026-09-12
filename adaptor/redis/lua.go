package redis

import "github.com/go-redis/redis"

var luaRenew = redis.NewScript(`
		local k = KEYS[1]
		local v = ARGV[1]
		if redis.call("GET", k) == v then
		  return redis.call("EXPIRE", k, ARGV[2])
		else
		  return 0
		end
`)

var luaUnlock = redis.NewScript(`
		local k = KEYS[1]
		local v = ARGV[1]
		if redis.call("GET", k) == v then
		  return redis.call("DEL", k)
		end
`)
