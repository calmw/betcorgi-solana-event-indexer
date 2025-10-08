package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/near/borsh-go"
)

// --- 配置 ---
const (
	programID = "CeEk3TZnjnbT71ojoSPUFw3z7t9wQ8sL3xwWNoecCKtb"
	rpcWS     = "wss://api.devnet.solana.com"
)

// ------------------ 事件结构 ------------------
type BetPlaced struct {
	GameID  uint64
	Player  [32]byte
	OrderID uint64
	Amount  uint64
	Hash    string
	Data    []byte
}

type AutoBetPlaced struct {
	Player      [32]byte
	OrderID     uint64
	GameID      uint64
	Amount      uint64
	ProfitStop  uint64
	LostStop    uint64
	AddWhenWin  uint8
	AddWhenLose uint8
	Hash        string
	Data        []byte
}

type WithdrawExecuted struct {
	Withdrawer [32]byte
	Amount     uint64
}

type AuthorizedPubkeyUpdated struct {
	OldPubkey [32]byte
	NewPubkey [32]byte
}

// ------------------ 注册机制 ------------------
type Event interface{}
type EventDecoder func([]byte) (Event, error)

var eventRegistry = map[string]EventDecoder{}

func registerEvent(name string, decoder EventDecoder) {
	d := sha256.Sum256([]byte("event:" + name))
	key := string(d[:8])
	eventRegistry[key] = decoder
	log.Printf("✅ 注册事件: %-25s discriminator=%x", name, d[:8])
}

func init() {
	registerEvent("BetPlaced", func(data []byte) (Event, error) {
		var e BetPlaced
		return &e, borsh.Deserialize(&e, data)
	})
	registerEvent("AutoBetPlaced", func(data []byte) (Event, error) {
		var e AutoBetPlaced
		return &e, borsh.Deserialize(&e, data)
	})
	registerEvent("WithdrawExecuted", func(data []byte) (Event, error) {
		var e WithdrawExecuted
		return &e, borsh.Deserialize(&e, data)
	})
	registerEvent("AuthorizedPubkeyUpdated", func(data []byte) (Event, error) {
		var e AuthorizedPubkeyUpdated
		return &e, borsh.Deserialize(&e, data)
	})
}

// ------------------ 主逻辑 ------------------
func main() {
	seen := make(map[string]struct{})

	for {
		ctx, cancel := context.WithCancel(context.Background())
		err := subscribeProgramLogs(ctx, rpcWS, programID, seen)
		if err != nil {
			log.Println("⚠️ 连接错误或断开，将在 5 秒后重连...", err)
			time.Sleep(5 * time.Second)
		}
		cancel()
	}
}

// ------------------ WebSocket 监听 ------------------
func subscribeProgramLogs(ctx context.Context, rpcWS, programID string, seen map[string]struct{}) error {
	conn, _, err := websocket.DefaultDialer.Dial(rpcWS, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "logsSubscribe",
		"params": []interface{}{
			map[string]interface{}{
				"mentions": []string{programID},
			},
			map[string]interface{}{
				"commitment": "finalized",
			},
		},
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		return err
	}

	log.Println("✅ 已订阅 program 日志，等待事件...")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		extractAndHandle(msg, seen)
	}
}

// ------------------ 工具函数 ------------------
func extractAndHandle(msg []byte, seen map[string]struct{}) {
	var rawMsg map[string]interface{}
	if err := json.Unmarshal(msg, &rawMsg); err != nil {
		return
	}

	params, ok := rawMsg["params"].(map[string]interface{})
	if !ok {
		return
	}
	result, ok := params["result"].(map[string]interface{})
	if !ok {
		return
	}
	value, ok := result["value"].(map[string]interface{})
	if !ok {
		return
	}
	logs, ok := value["logs"].([]interface{})
	if !ok {
		return
	}

	for _, l := range logs {
		line, ok := l.(string)
		if !ok {
			continue
		}
		const prefix = "Program data: "
		if len(line) > len(prefix) && line[:len(prefix)] == prefix {
			dataB64 := line[len(prefix):]
			handleProgramData(dataB64, seen)
		}
	}
}

func handleProgramData(dataB64 string, seen map[string]struct{}) {
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil || len(raw) <= 8 {
		return
	}

	key := dataB64
	if _, ok := seen[key]; ok {
		return
	}
	seen[key] = struct{}{}

	discriminator := string(raw[:8])
	payload := raw[8:]

	if decode, ok := eventRegistry[discriminator]; ok {
		event, err := decode(payload)
		if err != nil {
			log.Printf("⚠️ 事件解析失败 (%x): %v", discriminator, err)
			return
		}
		b, _ := json.MarshalIndent(event, "", "  ")
		log.Printf("🎯 事件 %x 解码成功:\n%s\n", discriminator, string(b))
	}
}
