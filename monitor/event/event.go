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

type DrawEvent struct {
	GameID      uint64
	Player      [32]byte
	OrderID     uint64
	Amount      uint64
	Seed        string
	HashExpired bool
}

// ------------------ 注册机制 ------------------

var Seen = map[string]struct{}{}

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
	registerEvent("DrawEvent", func(data []byte) (Event, error) {
		var e DrawEvent
		return &e, borsh.Deserialize(&e, data)
	})
}

// MarshalEvent ------------------ 工具 ------------------
func MarshalEvent(event Event) string {
	b, _ := json.MarshalIndent(event, "", "  ")
	return string(b)
}

// ExtractAndHandle ------------------ 工具函数 ------------------
func ExtractAndHandle(msg []byte) {
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
			HandleProgramData(dataB64, signature, slot)
		}
	}
}

func HandleProgramData(dataB64 string, signature string, slot uint64) {
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil || len(raw) <= 8 {
		return
	}

	discriminator := fmt.Sprintf("%x", raw[:8])
	key := fmt.Sprintf("%s:%x", signature, raw[:8])
	if _, ok := Seen[key]; ok {
		return
	}
	Seen[key] = struct{}{}

	payload := raw[8:]

	if decode, ok := EventRegistry[discriminator]; ok {
		event, err := decode(payload)
		if err != nil {
			log.Printf("⚠️ 事件解析失败 (%x): %v", discriminator, err)
			return
		}
		b, _ := json.MarshalIndent(event, "", "  ")
		log.Printf("🎯 事件 %x 解码成功:\n%s\n", discriminator, string(b)) // 保存到数据库
		log.Printf("discriminator: %x", discriminator)
		err = model.SaveEventToDB(signature, slot, discriminator, payload)
		if err != nil {
			log.Println("SaveEventToDB：", err)
			return
		}
		if discriminator == "585891e27ece2000" { // BetPlaced
			betEvent := event.(*BetPlaced)
			player := base58.Encode(betEvent.Player[:])
			err := model.SaveEventBetToDB(dataB64, betEvent.GameID, betEvent.OrderID, player, fmt.Sprintf("%d", betEvent.Amount), betEvent.Hash, string(betEvent.Data), signature)
			if err != nil {
				log.Println("SaveEventBetToDB：", err)
				return
			}
		} else if discriminator == "e86e28b168b9313b" { // DrawEvent
			drawEvent := event.(*DrawEvent)
			player := base58.Encode(drawEvent.Player[:])
			err := model.SaveEventDrawToDB(dataB64, drawEvent.GameID, drawEvent.OrderID, player, fmt.Sprintf("%d", drawEvent.Amount), drawEvent.Seed, signature, drawEvent.HashExpired)
			if err != nil {
				log.Println("SaveEventDrawToDB：", err)
				return
			}
		}
	} else {
		log.Printf("⚠️ 未知事件 discriminator: %x", discriminator)
	}
}
