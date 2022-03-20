package config

import "fmt"

type DBConfig struct {
	Name     string `mapstructure:"name"`
	Database string `mapstructure:"database"`
	Host     string `mapstructure:"addr"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func (conf *DBConfig) DSN() string {
	dsn := conf.Database
	if conf.Name == "postgres" {
		return fmt.Sprintf(
			"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Asia/Shanghai",
			conf.Host,
			conf.User,
			conf.Password,
			conf.Database,
			conf.Port)
	}

	if conf.Name == "mysql" {
		return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
			conf.User,
			conf.Password,
			conf.Host,
			conf.Port,
			conf.Database)

	}
	return dsn
}