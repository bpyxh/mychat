package service

import (
	"fmt"
	"time"
)

// TODO: 需要加锁
var onlineUser map[int64]map[string]interface{} = make(map[int64]map[string]interface{})

var onlineUserUpdated = true

// TODO: 没有变化不用发送
func broadOnlineUser() {
	ticker := time.NewTicker(3 * time.Second)

	// defer 确保在 main 函数结束时，Ticker 被安全关闭
	defer ticker.Stop()

	fmt.Println("Ticker 启动，每 500ms 触发一次...")

	// 2. 使用 for 循环配合 range Ticker 的通道 (C)
	for range ticker.C {
		// fmt.Println("online user updated, ", onlineUserUpdated)
		if !onlineUserUpdated {
			continue
		}
		// onlineUser := getOnlineUserCore()
		// data, _ := json.Marshal(onlineUser)
		resp := make(map[string]interface{})
		resp["to_id"] = -1.0
		resp["cmd"] = 3
		resp["user_info"] = getOnlineUser()
		// data, _ := json.Marshal(resp)

		BroadMsg(resp)

		onlineUserUpdated = false
		// clientMap[userId] = node
		// rwLocker.Unlock()
	}
}

func init() {
	go broadOnlineUser()
}

func AddOnlineUser(name string, userId int64) {
	fmt.Println("addonlineuser", name, userId)

	if _, ok := onlineUser[userId]; !ok {
		onlineUser[userId] = map[string]interface{}{
			"name":    name,
			"user_id": userId,
		}

		onlineUserUpdated = true
	}
}

// func RemoveOnlineUser(userId int64) {

// }

func getOnlineUser() []map[string]interface{} {
	userInfo := []map[string]interface{}{}
	for _, v := range onlineUser {
		userInfo = append(userInfo, v)
	}
	return userInfo
}

func IsUserOnline(name string) bool {
	for _, v := range onlineUser {
		if v["name"] == name {
			return true
		}
	}

	return false
}
