package restapi

import (
	"github.com/go-redis/redis/v7"
	"github.com/sonic-net/sonic-swss-common/goext/redisconn"
)

func newRedisClient(options *redis.Options) *redis.Client {
	options.Dialer = redisconn.WrapDialer(options.Dialer, options.DialTimeout, options.TLSConfig)
	return redis.NewClient(options)
}
