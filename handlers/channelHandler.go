package handlers

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

type ChannelHandler struct {
	client     *telegram.Client
	dispatcher *updates.Manager
	channels   []string                             // 要监听的频道用户名列表
	callback   func(channel string, message string) // 消息回调函数
}

func NewChannelHandler(client *telegram.Client, channels []string, callback func(string, string)) *ChannelHandler {
	dispatcher := updates.New(updates.Config{
		Handler: tg.NewUpdateDispatcher(),
	})
	return &ChannelHandler{
		client:     client,
		dispatcher: dispatcher,
		channels:   channels,
		callback:   callback,
	}
}
