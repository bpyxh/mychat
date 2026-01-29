package middleware

import (
	"errors"
	"mychat/internal/handler/dto"
	"mychat/internal/model"
	"strconv"
	"time"

	"go.uber.org/zap"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	IdentityKey = "currentUser"
	UserID      = "userId"
	UserName    = "userName"
	loginFailed = "loginFailed"
)

func GetContextUserID(ctx *gin.Context) (result uint64, err error) {
	if data, ok := ctx.Get(IdentityKey); ok {
		if userData, ok := data.(map[string]any); ok {
			if userId, ok := userData[UserID].(string); ok {
				temp, err := strconv.Atoi(userId)
				if err == nil {
					result = uint64(temp)
					return result, err
				}
			}
		}
	}

	return 0, errors.New("invalid context data")
}

func AuthMiddleWare() *jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "xd zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: IdentityKey,
		// 设置token里包含的信息
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(map[string]interface{}); ok {
				var userId, username string
				if userId, ok = v["user_id"].(string); !ok {
					zap.S().Errorf("invalid jwt payload data, %v", data)
					return jwt.MapClaims{}
				}
				if username, ok = v["username"].(string); !ok {
					zap.S().Errorf("invalid jwt payload data, %v", data)
					return jwt.MapClaims{}
				}

				return jwt.MapClaims{
					UserID:   userId,
					UserName: username,
				}
			}
			return jwt.MapClaims{}
		},
		// 根据解析token里的信息返回一个对象，GinJWTMiddleware内部会把这个值Set到gin.Context里
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return map[string]interface{}{
				UserID:   claims[UserID],
				UserName: claims[UserName],
			}
		},
		// 只有登录路由会用, 返回的数据会传给PayloadFunc
		Authenticator: func(c *gin.Context) (interface{}, error) {
			// u, ok := c.Get(IdentityKey)
			// zap.S().Debug("Authenticator u:", u, "OK:", ok)
			// if ok {
			// 	return u, nil
			// }
			var loginVals dto.LoginReq
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			user, err := model.Login(loginVals.UserName, loginVals.Password)
			if err != nil {
				c.Set(loginFailed, loginFailed)
				return nil, err
			}

			c.Set("user", user)

			tokenPayload := BuildTokenPayload(user.ID, user.UserName)
			return tokenPayload, nil
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			userData, ok := c.Get("user")
			if !ok {
				zap.S().Errorln("get user from context error.")
				c.JSON(400, dto.ErrorResp{
					Msg: "登录失败",
				})
				return
			}
			user, ok := userData.(*model.User)
			if !ok {
				zap.S().Errorln("convert to model user error.")
				c.JSON(400, dto.ErrorResp{
					Msg: "登录失败",
				})
				return
			}
			userResp := dto.UserResp{}
			userResp.ID = user.ID
			userResp.UserName = user.UserName
			userResp.Name = user.Name

			c.JSON(200, dto.LoginResp{
				Response: dto.Response{
					Code: 200,
					Msg:  "操作成功",
				},
				Data: dto.LoginRespData{
					Expire: expire.Format(time.RFC3339),
					Token:  token,
					User:   userResp,
				},
			})
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
			// allowedUsers := config.Config.AllowedUsers
			// if v, ok := data.(map[string]any); ok {
			// 	if userName, ok := v[UserName].(string); ok {
			// 		if len(allowedUsers) == 0 || utils.Contains(config.Config.AllowedUsers, userName) {
			// 			return true
			// 		}
			// 	}
			// }

			// return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			reason, ok := c.Get(loginFailed)
			if ok {
				reason, ok := reason.(string)
				if ok && reason == loginFailed {
					code = 400
				}
			}

			c.JSON(code, dto.ErrorResp{
				Msg: "认证失败, " + message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		zap.S().Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		zap.S().Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	return authMiddleware
}

func BuildTokenPayload(userId uint32, userName string) map[string]any {
	data := map[string]any{
		"user_id":  strconv.Itoa(int(userId)),
		"username": userName,
	}

	return data
}
