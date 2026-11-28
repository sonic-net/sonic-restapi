package restapi

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"
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
