package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type Config struct {
	DB      DBConfig      `mapstructure:"db"`      // DB 数据库连接配置
	Service ServerConfig  `mapstructure:"service"` // Service 服务配置
	Gateway GatewayConfig `mapstructure:"gateway"`
	Cert    CertConfig    `mapstructure:"sign"` // Cert JWT 使用RSA的密钥配置
}

func Read(section string, cfg interface{}) error {
	v := viper.GetViper()
	if section != "" {
		v = viper.Sub(section)
	}

	if v == nil {
		errStr := fmt.Sprintf("没有找到%s配置", section)
		glog.Error(errStr)
		return errors.New(errStr)
	}

	if err := v.Unmarshal(cfg); err != nil {
		glog.Fatalf("读取配置出错:%v", err)
		return err
	}
	return nil
}

func Init(conf interface{}) error {
	wd, _ := os.Getwd()
	viper.AddConfigPath(wd)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config.yaml")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if err = viper.Unmarshal(conf); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func InitRemote(provider, endpoint, configPath string, conf interface{}) error {
	runtime_viper := viper.New()
	if err := runtime_viper.AddRemoteProvider(provider, endpoint, configPath); err != nil {
		return err
	}

	// 因为在字节流中没有文件扩展名，所以这里需要设置下类型。
	// 支持的扩展名有 "json", "toml", "yaml", "yml", "properties",
	// "props", "prop", "env", "dotenv"
	runtime_viper.SetConfigType("yaml")

	if err := runtime_viper.ReadRemoteConfig(); err != nil {
		return err
	}

	// 反序列化
	runtime_viper.Unmarshal(&conf)
	runtime_viper.WatchRemoteConfigOnChannel()
	return nil
}
