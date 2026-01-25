package handler

import (
	"fmt"
	"mychat/internal/handler/dto"
	"mychat/internal/model"
	"net/url"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func register(ctx *gin.Context) {
	var req dto.CreateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zap.S().Info("register: ", err)

		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, f := range errs {
				if f.Tag() == "required" {
					ctx.JSON(400, dto.Response{
						Code: 100,
						Msg:  fmt.Sprintf("%s 不能为空", f.Field()),
					})
					return
				}
			}
		}

		ctx.JSON(400, dto.Response{
			Code: 100,
			Msg:  "请求参数错误",
		})
		return
	}

	if req.UserName == "" || req.Password2 == "" || req.Password2 == "" {
		ctx.JSON(400, dto.Response{
			Code: 100,
			Msg:  "请求参数错误",
		})
		return
	}

	if req.Password != req.Password2 {
		ctx.JSON(400, dto.Response{
			Code: 100,
			Msg:  "两次密码不一致",
		})
		return
	}

	form := url.Values{}
	form.Set("username", req.UserName)
	form.Set("email", req.Email)
	form.Set("password", req.Password2)
	form.Set("name", req.Name)
	err := model.CreateUser(form)
	if err != nil {
		ctx.JSON(400, dto.Response{
			Code: 1000,
			Msg:  err.Error(),
		})
		return
	}

	resp := map[string]any{
		"username": req.UserName,
		"email":    req.Email,
	}

	ctx.JSON(200, dto.Response{
		Code: 200,
		Msg:  "注册用户成功",
		Data: resp,
	})
}

// func DeleteUser(ctx *gin.Context) {

// }

// func InitWss(ctx *gin.Context) {

// }

// func GetTestUser(ctx *gin.Context) {
// 	userInfo := dao.GetTestUserInfo()
// 	for _, v := range userInfo {
// 		if !IsUserOnline(v[0]) {
// 			ctx.JSON(200, gin.H{
// 				"code":     0,
// 				"name":     v[0],
// 				"password": v[1],
// 			})
// 			return
// 		}
// 	}

// 	ctx.JSON(200, gin.H{
// 		"code": -1,
// 		"msg":  "测试用户都被占用了",
// 	})
// }

func InitUserRouter(g *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	g.POST("/user", register)
}
