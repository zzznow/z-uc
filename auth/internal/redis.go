package internal

import (
	"context"
	"fmt"
	LOGGER "log/slog"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *RedisConfig) (err error) {
	RDB = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port),
		Password: cfg.Passwd,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		LOGGER.Warn("redis ping failed", "error", err.Error())
	}
	return err
}

func CloseRedis() {
	if RDB != nil {
		_ = RDB.Close()
	}
}
