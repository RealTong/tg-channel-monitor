package main

import (
	"context"
	"crypto-message-benchmark/utils"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

type ENV struct {
	appID       int
	appHash     string
	sessionPath string
	phone       string
	wsPort      string
}

func loadENV() ENV {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	appID, err := strconv.Atoi(os.Getenv("APPID"))
	if err != nil {
		log.Fatal("加载 APPID 失败")
	}
	appHash := os.Getenv("APP_HASH")
	sessionPath := os.Getenv("SESSION_PATH")
	phone := os.Getenv("TG_PHONE")
	wsPort := os.Getenv("WS_PORT")

	if appHash == "" || sessionPath == "" {
		log.Fatal("加载环境变量失败")
	}
	e := ENV{
		appID:       appID,
		appHash:     appHash,
		sessionPath: sessionPath,
		phone:       phone,
		wsPort:      wsPort,
	}
	return e
}

func main() {
	// ctrl + c
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// load .env file
	env := loadENV()
	appID := env.appID
	appHash := env.appHash
	sessionPath := env.sessionPath
	phone := env.phone
	// wsPort := env.wsPort

	sessionStorage := &session.FileStorage{
		Path: sessionPath,
	}

	dispatcher := tg.NewUpdateDispatcher()
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: sessionStorage,
		UpdateHandler: updates.New(updates.Config{
			Handler: dispatcher,
		}),
	})

	// 启动 WebSocket 服务器
	// go func() {
	// 	log.Printf("启动 WebSocket 服务器，端口: %s", wsPort)
	// 	ws.Start(wsPort)
	// }()

	// // 启动 WebSocket 客户端
	// log.Printf("启动 WebSocket 客户端，连接到: %s", "wss://bwenews-api.bwe-ws.com/ws")
	// // 获取当前时间戳 毫秒级别
	// wsClient := ws.NewClient("wss://bwenews-api.bwe-ws.com/ws", func(message ws.Message) {
	// 	log.Printf("收到 WebSocket 消息：%+v", message)
	// 	currentTimems := time.Now().UnixNano() / 1e6
	// 	// 通过 Bot 发送 WebSocket 消息到 Telegram
	// 	messageText := "BWENews WebSocket 消息:\n类型：" + message.SourceName + "\n内容：" + message.NewsTitle + "\n时间戳：" + strconv.FormatInt(currentTimems, 10)
	// 	if err := utils.SendBotMessage(messageText); err != nil {
	// 		log.Printf("发送 WebSocket 消息到 Telegram 失败: %v", err)
	// 	} else {
	// 		log.Printf("WebSocket 消息已转发到 Telegram")
	// 	}
	// })

	// go func() {
	// 	if err := wsClient.Connect(ctx); err != nil {
	// 		log.Printf("WebSocket 连接失败: %v", err)
	// 	}
	// }()
	// defer wsClient.Close()

	// 启动 Telegram 客户端
	err := client.Run(ctx, func(ctx context.Context) error {
		// 检查认证状态
		status, _ := client.Auth().Status(ctx)
		log.Printf("Authorized: %+v\n", status.Authorized)

		// 如果未认证，执行认证流程
		if !status.Authorized {
			log.Printf("执行认证")
			flow := auth.NewFlow(utils.Terminal{PhoneNumber: phone}, auth.SendCodeOptions{})
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}
		}

		// 获取源频道 ID
		monitorChannels := []string{"BWENews", "news6551", "NewListingsFeed", "TrumpTruthSocial_Alert"}

		// 注册频道消息处理器
		dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
			msg, ok := update.Message.(*tg.Message)
			if !ok {
				return nil
			}

			// 检查是否是来自源频道的消息
			if peerChannel, ok := msg.PeerID.(*tg.PeerChannel); ok {
				for _, channel := range monitorChannels {
					sourceID := utils.GetIDFromDomain(ctx, client, channel)
					if peerChannel.ChannelID == sourceID {
						// 获取当前时间戳 毫秒级别
						currentTimems := time.Now().UnixNano() / 1e6
						messageText := msg.Message
						if messageText == "" {
							return nil
						}
						if err := utils.SendBotMessage(messageText + "\n时间戳：" + strconv.FormatInt(currentTimems, 10)); err != nil {
							log.Printf("转发消息失败: %v", err)
						} else {
							log.Printf("消息已成功转发")
						}
					}
				}
			}

			return nil
		})

		// 创建一个信号通道，用于阻塞主程序直到收到中断信号
		log.Println("开始监听消息，按 Ctrl+C 退出...")
		<-ctx.Done()
		return nil
	})

	if err != nil {
		log.Fatalf("运行失败: %v", err)
	}
}
