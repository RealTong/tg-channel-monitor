package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Message 定义 WebSocket 消息结构
type Message struct {
	SourceName    string   `json:"source_name"`
	NewsTitle     string   `json:"news_title"`
	CoinsIncluded []string `json:"coins_included"`
	Url           string   `json:"url"`
	Timestamp     int64    `json:"timestamp"`
}

// Client 是 WebSocket 客户端
type Client struct {
	URL             string
	MessageCallback func(message Message)
	conn            *websocket.Conn
	done            chan struct{}
}

// NewClient 创建新的 WebSocket 客户端
func NewClient(wsURL string, messageCallback func(message Message)) *Client {
	return &Client{
		URL:             wsURL,
		MessageCallback: messageCallback,
		done:            make(chan struct{}),
	}
}

// Connect 连接到 WebSocket 服务器
func (c *Client) Connect(ctx context.Context) error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	log.Printf("连接到 WebSocket 服务器: %s", c.URL)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn

	// 启动读取消息的 goroutine
	go c.readMessages(ctx)

	return nil
}

// readMessages 持续读取 WebSocket 消息
func (c *Client) readMessages(ctx context.Context) {
	defer c.Close() // 确保资源释放

	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket 客户端关闭")
			return
		default:
			// 读取消息
			_, msgBytes, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket 读取错误: %v", err)
				}
				// 尝试重新连接
				log.Println("尝试重新连接...")
				time.Sleep(5 * time.Second)
				if err := c.Connect(ctx); err != nil {
					log.Printf("重新连接失败: %v", err)
					return
				}
				continue
			}

			// 解析消息
			var message Message
			if err := json.Unmarshal(msgBytes, &message); err != nil {
				log.Printf("解析消息失败: %v", err)
				continue
			}

			// 处理消息
			if c.MessageCallback != nil {
				c.MessageCallback(message)
			}
		}
	}
}

// Close 关闭 WebSocket 连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		select {
		case <-c.done:
		case <-time.After(time.Second):
		}
		c.conn.Close()
	}
}
