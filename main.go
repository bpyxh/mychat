package main

import (
	"mychat/initialize"
	"mychat/router"
	"net/http"

	"github.com/gin-gonic/gin"
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

	router := router.Router()
	router.Run(":8800")
}

func Pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":   "xxx",
		"age":    18,
		"school": "nonono",
	})
}
