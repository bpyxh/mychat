package service

import (
	"fmt"
	"math/rand"
	"mychat/common"
	"mychat/dao"
	"mychat/middleware"
	"mychat/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func List(ctx *gin.Context) {
	list, err := dao.GetUserList()
	if err != nil {
		ctx.JSON(200, gin.H{
			"code":    -1,
			"message": "获取用户列表失败",
		})

		return
	}

	ctx.JSON(http.StatusOK, list)
}

func LoginByNameAndPassWord(ctx *gin.Context) {
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")
	data, err := dao.FindUserByName(name)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code":    -1,
			"message": "登录失败",
		})
		return
	}
	if data.Name == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户名不存在",
		})
	}
	ok := common.CheckPassWord(password, data.Salt, data.Password)
	if !ok {
		ctx.JSON(200, gin.H{
			"code":    -1,
			"message": "密码错误",
		})
	}
	Rsp, err := dao.FindUserByNameAndPwd(name, data.Password)
	if err != nil {
		zap.S().Info("登录失败", err)
		return
	}
	token, err := middleware.GenerateToken(Rsp.Id, "yk")
	if err != nil {
		zap.S().Info("生成token失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":   0,
		"msg":    "登录成功",
		"token":  token,
		"userId": Rsp.Id,
		"name":   Rsp.Name,
	})

	fmt.Println("before AddOnlineUser")
	AddOnlineUser(name, int64(data.Id))
}

func NewUser(ctx *gin.Context) {
	user := models.User{}
	user.Name = ctx.Request.FormValue("name")
	password := ctx.Request.FormValue("password")
	repassword := ctx.Request.FormValue("repassword")

	fmt.Println(user.Name, password, repassword)

	if user.Name == "" || password == "" || repassword == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户名或密码不能为空！",
			"data": user,
		})
		return
	}

	_, err := dao.FindUser(user.Name)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "该用户已注册",
			"data": user,
		})
		return
	}

	if password != repassword {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "两次密码不一致！",
			"data": user,
		})
		return
	}

	salt := fmt.Sprintf("%d", rand.Int31())
	user.Password = common.SaltPassWord(password, salt)
	user.Salt = salt
	t := time.Now()
	user.CreateAt = t
	user.UpdatedAt = t
	user.LoginTime = &t
	user.LoginOutTime = &t
	user.HeartBeatTime = &t
	dao.CreateUser(user)
	ctx.JSON(200, gin.H{
		"code": 0,
		"msg":  "新增用户成功！",
		"data": user,
	})
}

func UpdateUser(ctx *gin.Context) {
	// user := models.User{}

	// id, err := strconv.Atoi(ctx.Request.FormValue("id"))
	// if err != nil {
	// 	zap.S().Info("类型转换失败", err)
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{
	// 		"code": -1,
	// 		"msg":  "注销账号失败",
	// 	})
	// 	return
	// }

	return
	// user.ID = uint(id)
	// name := ctx.Request.FormValue("name")
	// password := ctx.Request.FormValue("password")
	// email := ctx.Request.FormValue("email")
	// phone := ctx.Request.FormValue("phone")
	// avatar := ctx.Request.FormValue("icon")
	// gender := ctx.Request.FormValue("gender")

}

func DeleteUser(ctx *gin.Context) {

}

func SendUserMsg(ctx *gin.Context) {
	chat(ctx.Writer, ctx.Request)
}

func InitWss(ctx *gin.Context) {

}

func GetTestUser(ctx *gin.Context) {
	userInfo := dao.GetTestUserInfo()
	for _, v := range userInfo {
		if !IsUserOnline(v[0]) {
			ctx.JSON(200, gin.H{
				"code":     0,
				"name":     v[0],
				"password": v[1],
			})
			return
		}
	}

	ctx.JSON(200, gin.H{
		"code": -1,
		"msg":  "测试用户都被占用了",
	})
}
