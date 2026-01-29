package db

import (
	"fmt"
	"mychat/internal/global"
	"mychat/internal/misc/config"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init() {
	// 启动时就打开数据库连接
	if err := connect(); err != nil {
		zap.S().Warn("mysql is not open:", err)
		panic(err)
	}
}

func connect() error {
	mysqlConfig := &config.Config.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.DbName,
		mysqlConfig.Charset)

	zap.S().Debug("dsn: ", dsn)

	var err error
	// TODO: 给gorm设置logger
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, err := global.DB.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdle)
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxConn)
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlConfig.MaxLifeTime) * time.Second)

	// 测试数据库连接是否 OK
	if err = sqlDB.Ping(); err != nil {
		panic(err)
	}

	return nil
}
