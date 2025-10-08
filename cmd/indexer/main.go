package main

import (
	"betcorgi-event-indexer/cmd/indexer/fetch"
	"betcorgi-event-indexer/cmd/indexer/ws"
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"log"
	"time"
)

func main() {

	db.InitMysql()
	//自动迁移
	err := db.DB.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").AutoMigrate(&model.Event{})
	if err != nil {
		log.Println("db AutoMigrate err: ", err)
	}

	go ws.ListenWS("wss://api.devnet.solana.com", "CeEk3TZnjnbT71ojoSPUFw3z7t9wQ8sL3xwWNoecCKtb")

	// 补漏，每30秒通过RPC检查一次
	for {
		fetch.FetchMissingEvents("https://api.devnet.solana.com", "CeEk3TZnjnbT71ojoSPUFw3z7t9wQ8sL3xwWNoecCKtb")
		time.Sleep(30 * time.Second)
	}
}
