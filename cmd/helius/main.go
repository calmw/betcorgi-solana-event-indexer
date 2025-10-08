package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/near/borsh-go"
)

// ------------------- Event Structs -------------------

// BetPlaced event
type BetPlaced struct {
	GameId  uint64
	Player  [32]byte
	OrderId uint64
	Amount  uint64
	Hash    string
	Data    []byte
}

// AuthorizedPubkeyUpdated event
type AuthorizedPubkeyUpdated struct {
	OldPubkey [32]byte
	NewPubkey [32]byte
}

// WithdrawExecuted event
type WithdrawExecuted struct {
	Withdrawer [32]byte
	Amount     uint64
}

// AutoBetPlaced event
type AutoBetPlaced struct {
	Player      [32]byte
	OrderId     uint64
	GameId      uint64
	Amount      uint64
	ProfitStop  uint64
	LostStop    uint64
	AddWhenWin  uint8
	AddWhenLose uint8
	Hash        string
	Data        []byte
}

// ------------------- Webhook Payload -------------------

type WebhookPayload struct {
	Transaction struct {
		Meta struct {
			LogMessages []string `json:"logMessages"`
		} `json:"meta"`
	} `json:"transaction"`
}

// ------------------- Helpers -------------------

func getDiscriminator(eventName string) []byte {
	// Anchor discriminator = sha256("event:<Name>")[0..8]
	preimage := []byte("event:" + eventName)
	hash := sha256.Sum256(preimage)
	return hash[0:8]
}

func hexDiscriminator(d []byte) string {
	return hex.EncodeToString(d)
}

// ------------------- Event Decoder -------------------

func decodeEvent(raw []byte) {
	if len(raw) < 8 {
		fmt.Println("Invalid event data")
		return
	}

	disc := raw[:8]
	payload := raw[8:]

	// 事件映射
	switch hexDiscriminator(disc) {
	case hexDiscriminator(getDiscriminator("BetPlaced")):
		var e BetPlaced
		if err := borsh.Deserialize(&e, payload); err == nil {
			fmt.Printf("🎲 BetPlaced: game_id=%d player=%x order_id=%d amount=%d hash=%s data=%x\n",
				e.GameId, e.Player, e.OrderId, e.Amount, e.Hash, e.Data)
		}
	case hexDiscriminator(getDiscriminator("AuthorizedPubkeyUpdated")):
		var e AuthorizedPubkeyUpdated
		if err := borsh.Deserialize(&e, payload); err == nil {
			fmt.Printf("🔑 AuthorizedPubkeyUpdated: old=%x new=%x\n", e.OldPubkey, e.NewPubkey)
		}
	case hexDiscriminator(getDiscriminator("WithdrawExecuted")):
		var e WithdrawExecuted
		if err := borsh.Deserialize(&e, payload); err == nil {
			fmt.Printf("💸 WithdrawExecuted: withdrawer=%x amount=%d\n", e.Withdrawer, e.Amount)
		}
	case hexDiscriminator(getDiscriminator("AutoBetPlaced")):
		var e AutoBetPlaced
		if err := borsh.Deserialize(&e, payload); err == nil {
			fmt.Printf("🤖 AutoBetPlaced: player=%x order_id=%d game_id=%d amount=%d profit_stop=%d lost_stop=%d add_win=%d add_lose=%d hash=%s data=%x\n",
				e.Player, e.OrderId, e.GameId, e.Amount, e.ProfitStop, e.LostStop, e.AddWhenWin, e.AddWhenLose, e.Hash, e.Data)
		}
	default:
		fmt.Printf("❓ Unknown event discriminator: %x\n", disc)
	}
}

// ------------------- Webhook Handler -------------------

func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	for _, log := range payload.Transaction.Meta.LogMessages {
		if len(log) > 25 && log[:25] == "Program log: Program data:" {
			base64Str := log[len("Program log: Program data: "):]
			raw, err := base64.StdEncoding.DecodeString(base64Str)
			if err != nil {
				fmt.Println("base64 decode error:", err)
				continue
			}
			decodeEvent(raw)
		}
	}

	w.WriteHeader(200)
}

// ------------------- Main -------------------

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, World!")
		if err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc("/webhook", handler)
	fmt.Println("Listening on :8080 ...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
}
