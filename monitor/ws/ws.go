package ws

import (
	event2 "betcorgi-event-indexer/monitor/event"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func ListenWS(rpcWS, programID string) {
	seen := map[string]struct{}{}

	for {
		conn, _, err := websocket.DefaultDialer.Dial(rpcWS, nil)
		if err != nil {
			log.Println("⚠️ WS连接失败，5秒后重试:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		subMsg := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "logsSubscribe",
			"params": []interface{}{
				map[string]interface{}{"mentions": []string{programID}},
				map[string]interface{}{"commitment": "finalized"},
			},
		}
		if err := conn.WriteJSON(subMsg); err != nil {
			log.Println("订阅失败:", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ 已订阅 program 日志，等待事件...")

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("⚠️ WS断开，重连中:", err)
				conn.Close()
				break
			}
			event2.ExtractAndHandle(msg, seen)
		}
	}
}
