package utils

import (
	"context"
	"fmt"
	"github.com/gotd/td/tg"
	"log"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
)

func getIDFromInputPeer(peer tg.InputPeerClass) (int64, error) {
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		return p.ChannelID, nil
	case *tg.InputPeerChat:
		return p.ChatID, nil
	case *tg.InputPeerUser:
		return p.UserID, nil
	default:
		return 0, fmt.Errorf("unknown peer type: %T", peer)
	}
}

func GetIDFromDomain(ctx context.Context, client *telegram.Client, domain string) int64 {
	// 判断第一个字符是不是@，如果是 去掉@
	if domain[0] == '@' {
		domain = domain[1:]
	}
	resolver := peer.DefaultResolver(client.API())
	peerID, err := resolver.ResolveDomain(context.Background(), domain)
	if err != nil {
		log.Printf("获取 id 错误: %v", err)
	}
	id, err := getIDFromInputPeer(peerID)
	return id
}
