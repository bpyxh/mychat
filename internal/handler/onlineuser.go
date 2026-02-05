package handler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var onlineUser map[uint64]UserInfo = make(map[uint64]UserInfo)
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
		// resp := make(map[string]interface{})
		// resp["to_id"] = -1.0
		// resp["cmd"] = 3
		// resp["user_info"] = getOnlineUser()
		// // data, _ := json.Marshal(resp)

		// broadOnlineUser()

		onlineUserUpdated.Store(false)
	}
}

func init() {
	onlineUserUpdated.Store(false)
	go broadOnlineUser()
}

func AddOnlineUser(name string, userId uint64) {
	fmt.Println("addonlineuser", name, userId)

	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := onlineUser[userId]; !ok {
		onlineUser[userId] = UserInfo{name, userId}

		onlineUserUpdated.Store(true)
	}
}

func RemoveOnlineUser(userId uint64) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := onlineUser[userId]; ok {
		delete(onlineUser, userId)
		onlineUserUpdated.Store(true)
	}
}

func getOnlineUser() (users []UserInfo) {
	mutex.RLock()
	defer mutex.RUnlock()
	for _, v := range onlineUser {
		users = append(users, v)
	}
	return users
}

func IsUserOnline(name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, v := range onlineUser {
		if v.Name == name {
			return true
		}
	}

	return false
}
