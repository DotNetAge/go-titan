package repository

import (
	"go-titan/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func WithPostgre(opts ...Option) *gorm.DB {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.Host == "" {
		options.Host = "127.0.0.1"
	}
	if options.Port == 0 {
		options.Port = 5432
	}
	options.Name = "postgres"
	db, err := gorm.Open(postgres.Open(options.DSN()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func WithConfig(conf *config.DBConfig) *gorm.DB {
	options := &Options{
		Name:     conf.Name,
		Database: conf.Database,
		Host:     conf.Host,
		Port:     conf.Port,
		User:     conf.User,
		Password: conf.Password,
	}
	if options.Name == "mysql" {
		db, err := gorm.Open(mysql.Open(options.DSN()), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		return db
	}

	if options.Name == "postgres" {
		db, err := gorm.Open(postgres.Open(options.DSN()), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		return db
	}

	db, err := gorm.Open(sqlite.Open(options.Database), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	return db
}

func WithMySQL(opts ...Option) *gorm.DB {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	options.Name = "mysql"
	if options.Host == "" {
		options.Host = "127.0.0.1"
	}
	if options.Port == 0 {
		options.Port = 3306
	}
	db, err := gorm.Open(mysql.Open(options.DSN()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func WithSQLite(fileName string) *gorm.DB {
	fn := fileName
	if fn == "" {
		fn = "data.db"
	}
	db, err := gorm.Open(sqlite.Open(fileName), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	return db
}
