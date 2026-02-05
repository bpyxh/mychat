package main

import (
	"fmt"
	"mychat/internal/db"
	"mychat/internal/handler"
	"mychat/internal/misc/config"
	"mychat/internal/misc/logger"
	"mychat/internal/router"
)

func main() {
	for range 3 {
		fmt.Println("----------------------------------------------------------------")
	}

	config.Init()

	logger := logger.Init()
	defer func() { _ = logger.Sync() }()

	db.Init()

	handler.InitClientManager()

	router.Run()
}
