package ws

import (
	"crypto-message-benchmark/utils"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 连接管理器
type ConnectionManager struct {
	connections map[*websocket.Conn]bool
	mutex       sync.Mutex
}

// 创建新的连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[*websocket.Conn]bool),
	}
}

// 添加连接
func (cm *ConnectionManager) AddConnection(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[conn] = true
	log.Printf("新的客户端连接，当前连接数：%d", len(cm.connections))
}

// 移除连接
func (cm *ConnectionManager) RemoveConnection(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.connections, conn)
	log.Printf("客户端断开连接，当前连接数：%d", len(cm.connections))
}

// 广播消息给所有连接
func (cm *ConnectionManager) Broadcast(message []byte) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	for conn := range cm.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("广播消息失败: %v", err)
			conn.Close()
			delete(cm.connections, conn)
		}
	}
}

// 消息请求结构
type MessageRequest struct {
	Message string `json:"message"` // 要发送的消息内容
	ChatID  string `json:"chat_id"` // 可选，目标聊天 ID
}

// 消息响应结构
type MessageResponse struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 响应消息
}

// 全局连接管理器
var connectionManager = NewConnectionManager()

// WebSocket 处理函数
func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	// 添加到连接管理器
	connectionManager.AddConnection(conn)

	// 确保连接关闭时移除
	defer func() {
		conn.Close()
		connectionManager.RemoveConnection(conn)
	}()

	// 发送欢迎消息
	welcomeResponse := MessageResponse{
		Success: true,
		Message: "已连接到 Telegram 消息服务",
	}
	welcomeJSON, _ := json.Marshal(welcomeResponse)
	conn.WriteMessage(websocket.TextMessage, welcomeJSON)

	// 持续读取消息
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 错误: %v", err)
			}
			return
		}

		// 只处理文本消息
		if messageType != websocket.TextMessage {
			continue
		}

		// 解析消息请求
		var request MessageRequest
		if err := json.Unmarshal(p, &request); err != nil {
			log.Printf("解析消息请求失败: %v", err)
			errorResponse := MessageResponse{
				Success: false,
				Message: "无效的消息格式",
			}
			errorJSON, _ := json.Marshal(errorResponse)
			conn.WriteMessage(websocket.TextMessage, errorJSON)
			continue
		}

		// 检查消息内容
		if request.Message == "" {
			errorResponse := MessageResponse{
				Success: false,
				Message: "消息内容不能为空",
			}
			errorJSON, _ := json.Marshal(errorResponse)
			conn.WriteMessage(websocket.TextMessage, errorJSON)
			continue
		}

		// 发送 Telegram 消息
		var sendErr error
		if request.ChatID != "" {
			// 如果指定了聊天 ID，发送到指定聊天
			sendErr = utils.SendBotMessageToChat(request.Message, request.ChatID)
		} else {
			// 否则使用默认聊天 ID
			sendErr = utils.SendBotMessage(request.Message)
		}

		// 发送响应
		var response MessageResponse
		if sendErr != nil {
			response = MessageResponse{
				Success: false,
				Message: "发送消息失败：" + sendErr.Error(),
			}
		} else {
			response = MessageResponse{
				Success: true,
				Message: "消息已发送",
			}
		}
		responseJSON, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, responseJSON)
	}
}

// Start 启动 WebSocket 服务器
func Start(port string) {
	if port == "" {
		port = "8081" // 默认端口
	}

	http.HandleFunc("/ws", handler)

	log.Printf("WebSocket 服务器启动在 :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动 WebSocket 服务器失败: %v", err)
	}
}
