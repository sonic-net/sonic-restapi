package restapi

import (
	"context"
	"net"

	"github.com/go-redis/redis/v7"
	"swsscommon/redisconn"
)

func newRedisClient(options *redis.Options) *redis.Client {
	if options.Network == "unix" && options.TLSConfig == nil {
		options.Dialer = func(ctx context.Context, _ string, address string) (net.Conn, error) {
			return redisconn.DialUnixContext(ctx, address, options.DialTimeout, true)
		}
	}
	return redis.NewClient(options)
}
