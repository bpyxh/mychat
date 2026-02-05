package handler

import (
	"encoding/json"
	"fmt"
	"mychat/internal/handler/dto"
	"mychat/internal/middleware"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
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

type UserInfo struct {
	Name   string `json:"name"`
	UserID uint64 `json:"user_id"`
}

type MsgCmd uint64

const (
	_ = iota
	TextMsg
	UserOnline
	GroupMsg
)

type Msg struct {
	Cmd      MsgCmd   `json:"cmd"`
	FromID   uint64   `json:"from_id"`
	ToID     uint64   `json:"to_Id"`
	Text     string   `json:"text"`
	UserInfo UserInfo `json:"user_info"`
}

type ClientManager struct {
	clients    map[uint64]*Client
	broadcast  chan *Msg
	register   chan *Client
	unregister chan *Client
}

func (h *ClientManager) run() {
	for {
		select {
		case client := <-h.register:
			uid := client.UserID()
			zap.S().Infof("%d online", uid)
			h.clients[uid] = client
		case client := <-h.unregister:
			uid := client.UserID()
			zap.S().Infof("%d offline", uid)
			if _, ok := h.clients[uid]; ok {
				delete(h.clients, uid)
				close(client.send)
			}

			RemoveOnlineUser(uid)
		case msg := <-h.broadcast:
			// 广播类型消息
			if msg.ToID == 0 {
				for _, client := range h.clients {
					data, _ := json.Marshal(msg)
					client.send <- data
				}
				continue
			}

			if _, ok := h.clients[msg.ToID]; ok {
				data, _ := json.Marshal(msg)
				h.clients[msg.ToID].send <- data
			} else {
				zap.S().Infof("Failed to find toId:%d client", msg.ToID)
			}
		}
	}
}

var clientManager *ClientManager

func newClientManager() *ClientManager {
	return &ClientManager{
		broadcast:  make(chan *Msg, 1000),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[uint64]*Client),
	}
}

func InitClientManager() {
	clientManager = newClientManager()
	go clientManager.run()
}

func BroadUserOnlineMsg(msg *Msg) {
	clientManager.broadcast <- msg
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
	userId        uint64
	send          chan []byte
	clientManager *ClientManager
}

func (c *Client) UserID() uint64 {
	return c.userId
}

// TODO: ping,pong 还没处理
func (c *Client) readPump() {
	defer func() {
		c.clientManager.unregister <- c
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

		msg := Msg{}
		err = json.Unmarshal(message, &msg)
		if err != nil {
			zap.S().Info("消息解析失败", err)
			return
		}
		// msg["self_id"] = node.userId

		c.processMsg(&msg)
	}
}

// TODO: 检查发送人的id和node记录的id是否同一个人
func (c *Client) processMsg(msg *Msg) {
	fmt.Println("processMsg... ", msg)

	switch msg.Cmd {
	case TextMsg:
		c.clientManager.broadcast <- msg
	case UserOnline:
		// resp := make(map[string]interface{})
		// resp["cmd"] = 3
		// resp["user_info"] = getOnlineUser()
		// data, _ := json.Marshal(resp)
		// c.send <- data

	case GroupMsg:
		fmt.Println("还没实现群发。")

	default:
		fmt.Printf("unknownd cmd: %#v\n", msg.Cmd)
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

func ws(ctx *gin.Context) {
	userID, err := middleware.GetContextUserID(ctx)
	if err != nil {
		zap.S().Info(err)
		ctx.JSON(500, dto.Response{
			Code: 100,
			Msg:  "内部错误",
		})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		zap.S().Error("upgrade ws error, %s", err)
		ctx.JSON(500, dto.Response{
			Code: 100,
			Msg:  "内部错误",
		})
		return
	}

	client := &Client{
		clientManager: clientManager,
		conn:          conn,
		userId:        userID, // TODO: 改成uint64
		send:          make(chan []byte, 256),
	}
	client.clientManager.register <- client

	zap.S().Infof("user %d connected ws", userID)

	go client.writePump()
	go client.readPump()
}

func InitWSRouter(g *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	g.GET("/ws", authMiddleware.MiddlewareFunc(), ws)
}
