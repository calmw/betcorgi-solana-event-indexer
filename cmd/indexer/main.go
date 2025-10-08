package main

import (
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"betcorgi-event-indexer/monitor/fetch"
	"betcorgi-event-indexer/monitor/ws"
	"log"
	"time"
)

func main() {

	db.InitMysql()
	//自动迁移
	err := db.DB.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").AutoMigrate(&model.Event{}, &model.EventBet{}, &model.EventDraw{})
	if err != nil {
		log.Println("db AutoMigrate err: ", err)
	}

	go ws.ListenWS("wss://api.devnet.solana.com", "FmoviYkRguNDJwX5vuj4NVPsy5Tzf9kZK23x6VTiE8P")

	// 补漏，每30秒通过RPC检查一次
	for {
		fetch.FetchMissingEvents("https://api.devnet.solana.com", "FmoviYkRguNDJwX5vuj4NVPsy5Tzf9kZK23x6VTiE8P")
		time.Sleep(30 * time.Second)
	}
}
