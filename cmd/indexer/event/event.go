package event

import (
	"betcorgi-event-indexer/model"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
)

// ------------------ 定义事件结构 ------------------

type Event interface{}

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

type EventDecoder func([]byte) (Event, error)

var EventRegistry = map[string]EventDecoder{}

func registerEvent(name string, decoder EventDecoder) {
	d := sha256.Sum256([]byte("event:" + name))
	key := fmt.Sprintf("%x", d[:8])
	EventRegistry[key] = decoder
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

// MarshalEvent ------------------ 工具 ------------------
func MarshalEvent(event Event) string {
	b, _ := json.MarshalIndent(event, "", "  ")
	return string(b)
}

// ------------------ 工具函数 ------------------
func ExtractAndHandle(msg []byte, seen map[string]struct{}) {
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
	context := result["context"].(map[string]interface{})
	slot := uint64(context["slot"].(float64))
	value, ok := result["value"].(map[string]interface{})
	if !ok {
		return
	}
	signature := value["signature"].(string)
	if signature == "1111111111111111111111111111111111111111111111111111111111111111" {
		log.Println("⚠️ 跳过虚拟 signature=1111... 的内部调用日志")
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
			handleProgramData(dataB64, seen, signature, slot)
		}
	}
}

func handleProgramData(dataB64 string, seen map[string]struct{}, signature string, slot uint64) {
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil || len(raw) <= 8 {
		return
	}

	discriminator := fmt.Sprintf("%x", raw[:8])
	key := fmt.Sprintf("%s:%x", signature, raw[:8])
	if _, ok := seen[key]; ok {
		return
	}
	seen[key] = struct{}{}

	payload := raw[8:]

	if decode, ok := EventRegistry[discriminator]; ok {
		event, err := decode(payload)
		if err != nil {
			log.Printf("⚠️ 事件解析失败 (%x): %v", discriminator, err)
			return
		}
		b, _ := json.MarshalIndent(event, "", "  ")
		log.Printf("🎯 事件 %x 解码成功:\n%s\n", discriminator, string(b)) // 保存到数据库
		playerStr := decodePlayer(event.(*BetPlaced))
		log.Printf("Player: %s", playerStr)
		log.Printf("discriminator: %x", discriminator)
		model.SaveEventToDB(signature, slot, discriminator, payload) // signature/slot实际可替换
	} else {
		log.Printf("⚠️ 未知事件 discriminator: %x", discriminator)
	}
}

func decodePlayer(event *BetPlaced) string {
	return base58.Encode(event.Player[:])
}
