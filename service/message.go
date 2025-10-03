package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gopkg.in/fatih/set.v0"
)

// Node 构造连接
type Node struct {
	Conn      *websocket.Conn //socket连接
	Addr      string          //客户端地址
	DataQueue chan []byte     //消息内容
	GroupSets set.Interface   //好友 / 群
	userId    int64
}

var clientMap map[int64]*Node = make(map[int64]*Node, 0)

var rwLocker sync.RWMutex

func chat(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	Id := query.Get("userId")
	userId, err := strconv.ParseInt(Id, 10, 64)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		return
	}

	var isvalid = true
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalid
		},
	}).Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("chat userId", userId)

	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
		userId:    userId,
	}

	fmt.Println("xx")

	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	// 服务发消息
	go sendMsgToClient(node)

	// 服务接收消息，从客户端接收消息，然后把消息发给目标人
	go recvProc(node)
}

// 从队列中取出自己的消息，然后发给自己的客户端
func sendMsgToClient(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				zap.S().Info("写入消息失败", err)
				return
			}
			fmt.Println("数据发送socket成功")
		}
	}
}

func recvProc(node *Node) {
	fmt.Println("h")
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			zap.S().Info("读取消息失败", err)
			return
		}

		fmt.Println("recv msg: ", string(data))

		msg := make(map[string]interface{})
		err = json.Unmarshal(data, &msg)
		if err != nil {
			zap.S().Info("消息解析失败", err)
			return
		}
		msg["self_id"] = node.userId

		allMsgChan <- msg
	}
}

var allMsgChan chan map[string]interface{} = make(chan map[string]interface{}, 1024)

func init() {
	go brodMsg()
	go broadOnlineUser()

}

// 服务端接收到消息，然后把消息分发出去。
func brodMsg() {
	for {
		select {
		case data := <-allMsgChan:
			fmt.Println("before dispatch msg")
			dispatch(data)
		}
	}
}

type UserInfo struct {
	Name   string
	UserId int64
}

func dispatch(msg map[string]interface{}) {
	// msg := Message{}
	// err := json.Unmarshal(data, &msg)
	// if err != nil {
	// 	zap.S().Info("消息解析失败", err)
	// 	return
	// }

	fmt.Println("dispatch... ", msg)

	cmd1, ok := msg["cmd"]
	// fmt.Println(cmd, ok)
	if !ok {
		zap.S().Info("invalid msg, cmd")
		return
	}
	cmd2, ok := cmd1.(float64)
	if !ok {
		zap.S().Info("invalid msg, cmd")
		return
	}

	cmd := int64(cmd2)
	zap.S().Info("cmd, ", cmd)

	switch cmd {
	case 1:
		sendMsg(msg)
	case 2:
		if val, ok := msg["self_id"].(int64); ok {
			getOnlineUser(val)
		} else {
			zap.S().Info("invalid msg, self_id")
		}

	case 3:
		fmt.Println("还没实现群发。")

	default:
		fmt.Printf("unknownd cmd: %#v\n", cmd)
	}
}

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
		resp["cmd"] = 3
		resp["user_info"] = getOnlineUserCore()
		data, _ := json.Marshal(resp)

		rwLocker.Lock()
		for _, v := range clientMap {
			v.Conn.WriteMessage(websocket.TextMessage, data)
		}
		rwLocker.Unlock()

		onlineUserUpdated = false
		// clientMap[userId] = node
		// rwLocker.Unlock()
	}
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

func RemoveOnlineUser(userId int64) {

}

func getOnlineUserCore() []map[string]interface{} {
	userInfo := []map[string]interface{}{}
	for _, v := range onlineUser {
		userInfo = append(userInfo, v)
	}
	return userInfo
}

// TODO: 跟定时器那个合并一下
func getOnlineUser(userId int64) {
	fmt.Println("getonlineuser", userId)

	rwLocker.Lock()
	node, ok := clientMap[userId]
	rwLocker.Unlock()
	if !ok {
		zap.S().Info("userID没有对应的node")
		return
	}
	resp := make(map[string]interface{})
	resp["cmd"] = 3
	resp["user_info"] = getOnlineUserCore()
	data, _ := json.Marshal(resp)
	node.DataQueue <- data
}

// TODO: 检查发送人的id和node记录的id是否同一个人
// 把消息发送给目标人
func sendMsg(msg map[string]interface{}) {
	zap.S().Debug("sendMsg msg, ", msg)

	toIdFloat, ok := msg["to_id"].(float64)
	if !ok {
		zap.S().Info("invalid msg, to_id")
		return
	}
	toId := int64(toIdFloat)

	rwLocker.Lock()
	node, ok := clientMap[int64(toId)]
	rwLocker.Unlock()

	if !ok {
		zap.S().Info("userID没有对应的node")
		return
	}
	zap.S().Info("to_id:", toId, "node:", node)

	if ok {
		msg["self_id"] = nil
		toMsg, _ := json.Marshal(msg)
		node.DataQueue <- []byte(toMsg)
	}
}
