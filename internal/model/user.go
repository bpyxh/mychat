package model

import (
	"errors"
	"fmt"
	"math/rand"
	"mychat/internal/global"
	"mychat/internal/utils"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uint32
	UserName string `gorm:"column:username" validate:"omitempty,validusername"`
	Name     string `validate:"omitempty,max=50"`
	Email    string `validate:"omitempty,email"`
	Password string `validate:"omitempty,min=3,max=100"`
	Salt     string
	Phone    string `validate:"omitempty,number,min=5,max=15"`

	LoginTime  time.Time `gorm:"column:login_time"`
	LogoutTime time.Time `gorm:"column:logout_time"`
}

func (table *User) TableName() string {
	return "user"
}

func validateUser(user *User) (errMsg string, err error) {
	if err = validate.Struct(user); err != nil {
		zap.S().Errorf("validate user error:%v", err)

		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
				field := e.Field()
				if field == "UserName" {
					errMsg = "用户名不合法!"
				} else if field == "Email" {
					errMsg = "邮箱不合法！"
				} else if field == "Password" {
					errMsg = "密码不合法！"
				} else {
					errMsg = "内部未知错误"
				}
			}
			err = errors.New(errMsg)
			return
		}
	}

	return
}

func CreateUser(form url.Values) (err error) {
	user, err := FindUser("email", form.Get("email"))
	if err != nil {
		err = errors.New("内部错误")
		return
	}
	if user != nil {
		err = errors.New("该邮箱已注册过")
		return
	}

	user, err = FindUser("username", form.Get("username"))
	if err != nil {
		err = errors.New("内部错误")
		return
	}
	if user != nil {
		err = errors.New("用户名已经存在")
		return
	}

	user = &User{}
	err = schemaDecoder.Decode(user, form)
	if err != nil {
		zap.S().Errorln("user schema Decode error:", err)
		err = errors.New("参数错误")
		return
	}

	_, err = validateUser(user)
	if err != nil {
		return err
	}

	salt := fmt.Sprintf("%d", rand.Int31())
	user.Password = utils.SaltPassWord(user.Password, salt)
	user.Salt = salt
	t := time.Now()
	user.CreatedAt = t
	user.LoginTime = t
	user.LogoutTime = t
	err = InsertUser(user)
	if err != nil {
		err = errors.New("内部服务错误")
		return
	}

	return nil
}

func InsertUser(user *User) error {
	tx := global.DB.Create(user)
	if tx.Error != nil {
		zap.S().Errorln("db error: ", tx.Error)
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		zap.S().Error("Failed to create user")
		return errors.New("新增用户失败")
	}

	return nil
}

func FindUser(field, val string) (*User, error) {
	user := User{}
	tx := global.DB.Where(field+" = ?", val).First(&user)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		zap.S().Errorln("db error: ", tx.Error)
		return nil, tx.Error
	}

	return &user, nil
}

func Login(username, passwd string) (*User, error) {
	var user User
	if username == "" {
		zap.S().Infof("username is empty")
		return nil, errors.New("用户名不能为空")
	}
	tx := global.DB.Where("username=? OR email=?", username, username).First(&user)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号或密码错误")
		}

		zap.S().Errorln("db error: ", tx.Error)
		return nil, tx.Error
	}

	if user.ID == 0 {
		zap.S().Infof("user %q is not exists!\n", username)
		return nil, errors.New("账号或密码错误")
	}

	if !utils.CheckPassWord(passwd, user.Salt, user.Password) {
		zap.S().Infof("用户名 %q 填写的密码错误\n", username)
		return nil, errors.New("账号或密码错误")
	}

	return &user, nil
}
