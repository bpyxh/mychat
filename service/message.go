package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
// newline = []byte{'\n'}
// space   = []byte{' '}
)

type Hub struct {
	clients    map[int64]*Client
	broadcast  chan map[string]interface{}
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan map[string]interface{}, 1000),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			uid := client.UserId()
			zap.S().Infof("%d register", uid)
			h.clients[uid] = client
		case client := <-h.unregister:
			uid := client.UserId()
			zap.S().Infof("%d ungisger", uid)
			if _, ok := h.clients[uid]; ok {
				delete(h.clients, uid)
				close(client.send)
			}

			RemoveOnlineUser(uid)
		case msg := <-h.broadcast:
			toIdFloat, ok := msg["to_id"].(float64)
			if !ok {
				zap.S().Info("invalid msg, to_id")
				return
			}
			toId := int64(toIdFloat)
			// 广播类型消息
			if toId == -1 {
				for _, client := range h.clients {
					data, _ := json.Marshal(msg)
					client.send <- data
				}
				continue
			}

			if _, ok := h.clients[toId]; ok {
				data, _ := json.Marshal(msg)
				h.clients[toId].send <- data
			} else {
				zap.S().Infof("Failed to find toId:%d client", toId)
			}
		}
	}
}

var hub *Hub

func init() {
	hub = newHub()
	go hub.run()
}

func BroadMsg(msg map[string]interface{}) {
	hub.broadcast <- msg
}

// Upgrader 配置，重要：设置 CheckOrigin 防止 CSWSH
var upgrader = websocket.Upgrader{
	// 生产环境中，必须设置合适的 Read/Write BufferSize
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 【最佳实践】在生产环境应验证 Origin 头部以防止 CSWSH 攻击
	CheckOrigin: func(r *http.Request) bool {
		// 允许来自特定域名的请求
		// return r.Header.Get("Origin") == "http://yourdomain.com"
		// 暂时允许所有来源：
		return true
	},
}

type Client struct {
	conn *websocket.Conn //socket连接
	// addr   string          //客户端地址
	userId int64
	send   chan []byte
	hub    *Hub
}

func (c *Client) UserId() int64 {
	return c.userId
}

// TODO: ping,pong 还没处理
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.S().Info("ws unexpected close error: %v", err)
			}
			break
		}
		fmt.Println("recv msg: ", string(message))

		msg := make(map[string]interface{})
		err = json.Unmarshal(message, &msg)
		if err != nil {
			zap.S().Info("消息解析失败", err)
			return
		}
		// msg["self_id"] = node.userId

		c.processMsg(msg)
	}
}

// TODO: 检查发送人的id和node记录的id是否同一个人
func (c *Client) processMsg(msg map[string]interface{}) {
	fmt.Println("processMsg... ", msg)

	cmdFloat, ok := msg["cmd"].(float64)
	if !ok {
		zap.S().Info("invalid msg, cmd")
		return
	}

	cmd := int64(cmdFloat)
	zap.S().Info("cmd, ", cmd)

	switch cmd {
	case 1:
		c.hub.broadcast <- msg
	case 2:
		resp := make(map[string]interface{})
		resp["cmd"] = 3
		resp["user_info"] = getOnlineUser()
		data, _ := json.Marshal(resp)
		c.send <- data

	case 3:
		fmt.Println("还没实现群发。")

	default:
		fmt.Printf("unknownd cmd: %#v\n", cmd)
	}
}

// TODO: 看一下客户端网页刷新后，ws是哪个代码路径断开的链接的。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				// w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func Ws(ctx *gin.Context) {
	w := ctx.Writer
	r := ctx.Request
	query := r.URL.Query()
	id := query.Get("userId")
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		return
	}

	// 升级 HTTP 连接到 WebSocket 协议
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().Info(err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		userId: userId,
		send:   make(chan []byte, 256),
	}
	client.hub.register <- client

	fmt.Println("chat userId", userId)

	go client.writePump()
	go client.readPump()
}
