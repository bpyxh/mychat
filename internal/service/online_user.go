package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var onlineUser map[int64]map[string]interface{} = make(map[int64]map[string]interface{})
var onlineUserUpdated atomic.Bool
var mutex sync.RWMutex

func broadOnlineUser() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		updated := onlineUserUpdated.Load()
		if !updated {
			continue
		}
		// data, _ := json.Marshal(onlineUser)
		resp := make(map[string]interface{})
		resp["to_id"] = -1.0
		resp["cmd"] = 3
		resp["user_info"] = getOnlineUser()
		// data, _ := json.Marshal(resp)

		BroadMsg(resp)

		onlineUserUpdated.Store(false)
	}
}

func init() {
	onlineUserUpdated.Store(false)
	go broadOnlineUser()
}

func AddOnlineUser(name string, userId int64) {
	fmt.Println("addonlineuser", name, userId)

	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := onlineUser[userId]; !ok {
		onlineUser[userId] = map[string]interface{}{
			"name":    name,
			"user_id": userId,
		}

		onlineUserUpdated.Store(true)
	}
}

func RemoveOnlineUser(userId int64) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := onlineUser[userId]; ok {
		delete(onlineUser, userId)
		onlineUserUpdated.Store(true)
	}
}

func getOnlineUser() []map[string]interface{} {
	userInfo := []map[string]interface{}{}
	mutex.RLock()
	defer mutex.RUnlock()
	for _, v := range onlineUser {
		userInfo = append(userInfo, v)
	}
	return userInfo
}

func IsUserOnline(name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, v := range onlineUser {
		if v["name"] == name {
			return true
		}
	}

	return false
}
