package initialize

import (
	"log"

	"go.uber.org/zap"
)

// TODO: 日志配置的应该更具体
func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("日志初始化失败", err.Error())
	}

	zap.ReplaceGlobals(logger)
}
