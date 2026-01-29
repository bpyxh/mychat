package router

import (
	"log"
	"mychat/internal/handler"
	"mychat/internal/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 跨域
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Cas-Ticket, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// @Summary 用户登录
// @Description 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @param user body dto.LoginReq true "登录信息"
// @Success 200 {object} dto.LoginResp "操作成功"
// @Failure 400 {object} dto.ErrorResp "操作失败"
// @Router /api/login [post]
func loginHandler(c *gin.Context) {
	authMiddleware := middleware.AuthMiddleWare()
	authMiddleware.LoginHandler(c)
}

func Run() {
	router := gin.Default()

	authMiddleware := middleware.AuthMiddleWare()

	router.Use(CORSMiddleware())

	router.NoRoute(func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	apiGroup := router.Group("/api")
	apiGroup.POST("/login", loginHandler)

	{
		apiGroup.GET("/refresh_token", authMiddleware.RefreshHandler)
		handler.InitUserRouter(apiGroup, authMiddleware)
		handler.InitWSRouter(apiGroup, authMiddleware)
	}

	// docs.SwaggerInfo.BasePath = "/"
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("服务器启动在 :8080")
	if err := router.Run(); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
