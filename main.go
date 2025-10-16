package main

import (
	"fmt"
	"mychat/dao"
	"mychat/initialize"
	"mychat/router"
)

func main() {
	initialize.InitLogger()

	// initialize.InitConfig()

	user := "root"
	password := "123456"
	host := "127.0.0.1"
	dbName := "mychat"
	port := 3306
	initialize.InitDB(user, password, host, dbName, port)

	dao.InitTestUser()

	fmt.Println("server running...................................................")

	router := router.Router()
	router.Run(":8800")
}
