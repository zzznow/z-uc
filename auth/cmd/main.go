package main

import (
	"fmt"
	"os"

	"github.com/zzznow/z-uc/auth"
	"github.com/zzznow/z-uc/auth/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "test"
	}

	if err := internal.InitConfig(env); err != nil {
		panic(err)
	}

	if err := internal.InitMysql(internal.Conf.MysqlConfig); err != nil {
		panic(err)
	}
	defer internal.CloseMysql()

	if err := internal.InitRedis(internal.Conf.RedisConfig); err != nil {
		fmt.Printf("warn: redis init failed: %v\n", err)
	} else {
		defer internal.CloseRedis()
	}

	gin.SetMode(internal.Conf.Mode)
	r := gin.Default()
	auth.RegisterRoutes(r)

	addr := fmt.Sprintf("%s:%d", internal.Conf.Host, internal.Conf.Port)
	fmt.Printf("Auth Service started at %s\n", addr)
	r.Run(addr)
}
