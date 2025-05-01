package utils

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// SendBotMessage 发送消息到默认聊天
func SendBotMessage(message string) error {
	chatID := os.Getenv("TARGET_CHAT_ID")


	return SendBotMessageToChat(message, chatID)
}

// SendBotMessageToChat 发送消息到指定聊天
func SendBotMessageToChat(message string, chatID string) error {
	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		return errors.New("BOT_TOKEN 环境变量未设置")
	}

	url := "https://api.telegram.org/bot" + botToken + "/sendMessage"

	// 检查 chatID 是否已经包含负号
	if !strings.HasPrefix(chatID, "-") && !strings.Contains(chatID, "@") {
		// 如果不是用户名（不包含@）且不以负号开头，则添加负号（群组 ID 通常为负数）
		chatID = "-" + chatID
	}

	// 构建请求体
	reqBody := strings.NewReader(`{"chat_id":"` + chatID + `","text":"` + message + `"}`)

	// 发送请求
	resp, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("发送消息失败，状态码：" + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
