package repository

import (
	"database/sql"
	"fmt"
)

type Option func(*Options)

type Options struct {
	Name     string `mapstructure:"name" yaml:"name"`
	Database string `mapstructure:"database" yaml:"database"`
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
}

func (opts *Options) Conn() (*sql.DB, error) {
	return sql.Open(opts.Name, opts.DSN())
}

func (opts *Options) DSN() string {
	dsn := opts.Database
	if opts.Name == "postgres" {
		return fmt.Sprintf(
			"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Asia/Shanghai",
			opts.Host,
			opts.User,
			opts.Password,
			opts.Database,
			opts.Port)
	}

	if opts.Name == "mysql" {
		return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
			opts.User,
			opts.Password,
			opts.Host,
			opts.Port,
			opts.Database)

	}
	return dsn
}

func Database(name string) Option {
	return func(opts *Options) {
		if len(name)>0 {
			opts.Database = name
		}
	}
}

func Port(port int) Option {
	return func(options *Options) {
		if port >0 {
			options.Port = port
		}
	}
}

func Host(host string) Option {
	return func(options *Options) {
		if host!="" {
			options.Host = host
		}
	}
}

func UserName(user string) Option {
	return func(options *Options) {
		if len(user)>0 {
			options.User = user
		}
	}
}

func Password(pwd string) Option {
	return func(options *Options) {
		options.Password = pwd
	}
}
