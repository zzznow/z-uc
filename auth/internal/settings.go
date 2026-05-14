package internal

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"`
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`

	BaseURL      string `mapstructure:"base_url"`
	*MysqlConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
	Apps         []WxmTokenEntry `mapstructure:"apps"`
}

type WxmTokenEntry struct {
	Id       string `mapstructure:"id"`
	TokenUrl string `mapstructure:"token_url"`
}

type MysqlConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Passwd       string `mapstructure:"passwd"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Passwd   string `mapstructure:"passwd"`
	PoolSize int    `mapstructure:"pool_size"`
}

func InitConfig(env string) (err error) {
	path := "config/application-" + env + ".yml"
	viper.SetConfigFile(path)
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.BindEnv("mysql.passwd")
	viper.BindEnv("redis.passwd")

	// apps.yml from ConfigMap mount — merge on top
	appsViper := viper.New()
	appsViper.SetConfigFile("config/apps.yml")
	appsViper.AddConfigPath(".")
	if mergeErr := appsViper.ReadInConfig(); mergeErr != nil {
		fmt.Printf("warn: apps.yml not found, apps list empty: %v\n", mergeErr)
	} else {
		for k, v := range appsViper.AllSettings() {
			viper.Set(k, v)
		}
		appsViper.WatchConfig()
		appsViper.OnConfigChange(func(in fsnotify.Event) {
			fmt.Println("apps.yml changed, reloading...")
			for k, v := range appsViper.AllSettings() {
				viper.Set(k, v)
			}
			if err = viper.Unmarshal(Conf); err != nil {
				fmt.Printf("viper.Unmarshal(Conf) failed, err:%v\n", err)
			}
		})
	}

	if err = viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal(Conf) failed, err:%v\n", err)
		return
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed...")
		if err = viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal(Conf) failed, err:%v\n", err)
		}
	})
	return
}
