package router

import (
	"mychat/middleware"
	"mychat/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("v1")
	user := v1.Group("user")
	{
		user.GET("/list", middleware.JWY(), service.List)
		user.POST("/login", service.LoginByNameAndPassWord)
		user.POST("/new", service.NewUser)
		user.DELETE("/delete", middleware.JWY(), service.DeleteUser)
		user.POST("/update", middleware.JWY(), service.UpdateUser)
		user.GET("/test_user", service.GetTestUser)
	}

	router.GET("/send_msg", middleware.JWY(), service.SendUserMsg)

	relation := v1.Group("relation").Use(middleware.JWY())
	{
		relation.POST("/list", service.FriendList)
		relation.POST("/add", service.AddFriendByName)
	}

	upload := v1.Group("upload")
	{
		upload.POST("/image", service.Image)
	}

	router.Static("/static", "./static")
	router.LoadHTMLGlob("view/*")

	router.GET("/", func(c *gin.Context) {
		// 使用 c.HTML() 渲染模板
		// 参数依次是：HTTP 状态码, 模板文件名, 传递给模板的数据 (gin.H 是 map[string]interface{} 的别名)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Welcome",
		})
	})

	return router
}
