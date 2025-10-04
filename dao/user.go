package dao

import (
	"errors"
	"fmt"
	"math/rand"
	"mychat/common"
	"mychat/global"
	"mychat/models"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func GetUserList() ([]*models.User, error) {
	var list []*models.User
	if err := global.DB.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("获取用户列表失败 %s", err)
	}

	return list, nil
}

func FindUserByNameAndPwd(name string, password string) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where("name = ? and password=?", name, password).First(&user); tx.RowsAffected == 0 {
		return nil, errors.New("未查询到记录")
	}

	t := strconv.Itoa(int(time.Now().Unix()))

	temp := common.Md5encoder(t)

	if tx := global.DB.Model(&user).Where("id = ?", user.Id).Update("identity", temp); tx.RowsAffected == 0 {
		return nil, errors.New("写入identity失败")
	}

	return &user, nil
}

func FindUserByName(name string) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where("name = ?", name).First(&user); tx.RowsAffected == 0 {
		return nil, errors.New("没有查询到")
	}

	return &user, nil
}

// 这个函数可能有问题。
func FindUser(name string) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where("name = ?", name).First(&user); tx.RowsAffected == 1 {
		return nil, errors.New("当前用户名已存在")
	}

	return &user, nil
}

func FindUserID(ID uint) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where(ID).First(&user); tx.RowsAffected == 0 {
		return nil, errors.New("为查询到记录")
	}
	return &user, nil
}

func FindUserByPhone(phone string) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where("phone = ?", phone).First(&user); tx.RowsAffected == 0 {
		return nil, errors.New("未查询到记录")
	}
	return &user, nil
}

func FindUerByEmail(email string) (*models.User, error) {
	user := models.User{}
	if tx := global.DB.Where("email = ?", email).First(&user); tx.RowsAffected == 0 {
		return nil, errors.New("未查询到记录")
	}
	return &user, nil
}

func CreateUser(user models.User) (*models.User, error) {
	tx := global.DB.Create(&user)
	if tx.RowsAffected == 0 {
		zap.S().Error("Failed to create user")
		return nil, errors.New("新增用户失败")
	}
	return &user, nil
}

func UpdateUser(user models.User) (*models.User, error) {
	tx := global.DB.Model(&user).Updates(models.User{
		Name:     user.Name,
		Password: user.Password,
		Gender:   user.Gender,
		Phone:    user.Phone,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Salt:     user.Salt,
	})

	if tx.RowsAffected == 0 {
		zap.S().Info("Failed to update user")
		return nil, errors.New("Failed to update user")
	}

	return &user, nil
}

func DeleteUser(user models.User) error {
	if tx := global.DB.Delete(&user); tx.RowsAffected == 0 {
		msg := "Failed to delete user"
		zap.S().Info(msg)
		return errors.New(msg)
	}
	return nil
}

func FriendList(userId uint) (*[]models.User, error) {
	relation := make([]models.Relation, 0)
	if tx := global.DB.Where("self_id = ? and type = 1", userId).Find(&relation); tx.RowsAffected == 0 {
		zap.S().Info("Failed to find relation data")
		return nil, errors.New("未查到好友关系")
	}
	userID := make([]uint, 0)
	for _, v := range relation {
		userID = append(userID, v.OtherId)
	}

	user := make([]models.User, 0)
	if tx := global.DB.Where("id in ?", userID).Find(&user); tx.RowsAffected == 0 {
		zap.S().Info("Failed to find releation friend")
		return nil, errors.New("未查到好友")
	}

	return &user, nil
}

func AddFriend(selfId, otherId uint) (int, error) {
	if selfId == otherId {
		return -2, errors.New("selfId 和 otherId相等")
	}

	otherUser, err := FindUserID(otherId)
	if err != nil {
		return -1, errors.New("未查询到用户")
	}
	if otherUser.Id == 0 {
		zap.S().Info("未查询到用户")
		return -1, errors.New("未查询到用户")
	}
	relation := models.Relation{}
	if tx := global.DB.Where("owner_id = ? and target_id = ? and type = 1", selfId,
		otherId).First(&relation); tx.RowsAffected == 1 {
		zap.S().Info("该好友存在")
		return 0, errors.New("好友已经存在")
	}
	if tx := global.DB.Where("owner_id = ? and target_id = ? and type = 1", otherId,
		selfId).First(&relation); tx.RowsAffected == 1 {
		zap.S().Info("该好友存在")
		return 0, errors.New("好友已经存在")
	}

	tx := global.DB.Begin()

	relation.SelfId = selfId
	relation.OtherId = otherUser.Id
	relation.Type = 1

	if t := tx.Create(&relation); t.RowsAffected == 0 {
		zap.S().Info("创建失败")
		tx.Rollback()
		return -1, errors.New("创建好友记录失败")
	}

	relation = models.Relation{}
	relation.SelfId = selfId
	relation.OtherId = otherId
	relation.Type = 1

	if t := tx.Create(&relation); t.RowsAffected == 0 {
		zap.S().Info("创建失败")
		tx.Rollback()
		return -1, errors.New("创建好友记录失败")
	}

	tx.Commit()
	return 1, nil
}

func AddFriendByName(userId uint, targetName string) (int, error) {
	user, err := FindUserByName(targetName)
	if err != nil {
		return -1, errors.New("该用户不存在")
	}
	if user.Id == 0 {
		zap.S().Info("未查询到该用户")
		return -1, errors.New("该用户不存在")
	}
	return AddFriend(userId, user.Id)
}

func createUser(name, password string) {
	user := models.User{}
	user.Name = name
	salt := fmt.Sprintf("%d", rand.Int31())
	user.Password = common.SaltPassWord(password, salt)
	user.Salt = salt
	t := time.Now()
	user.CreateAt = t
	user.UpdatedAt = t
	user.LoginTime = &t
	user.LoginOutTime = &t
	user.HeartBeatTime = &t
	CreateUser(user)
}

func GetTestUserInfo() [][]string {
	return [][]string{
		{"zsan", "lll"},
		{"lisi", "lll"},
		{"wwu", "lll"},
		{"xwang", "lll"},
		{"wzhao", "lll"},
	}
}

func InitTestUser() {
	users, err := GetUserList()
	if err != nil {
		panic(fmt.Sprintf("failed to get user list, %s", err))
	}

	if len(users) > 0 {
		return
	}

	userInfo := GetTestUserInfo()
	for _, v := range userInfo {
		createUser(v[0], v[1])
	}
}
