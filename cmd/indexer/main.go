package main

import (
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"betcorgi-event-indexer/monitor/fetch"
	"betcorgi-event-indexer/monitor/ws"
	"log"
	"os"
	"time"
)

func main() {

	db.InitMysql()
	//自动迁移
	err := db.DB.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").AutoMigrate(&model.Event{}, &model.EventBet{}, &model.EventDraw{})
	if err != nil {
		log.Println("db AutoMigrate err: ", err)
	}
	RpcWs := os.Getenv("RPC_WS")
	RpcUrl := os.Getenv("RPC_URL")
	ProgramId := os.Getenv("PROGRAM_ID")
	log.Println("RPC_WS", RpcWs)
	log.Println("RPC_URL", RpcUrl)
	log.Println("PROGRAM_ID", ProgramId)

	go ws.ListenWS(RpcWs, ProgramId)

	// 补漏，每30秒通过RPC检查一次
	for {
		fetch.FetchMissingEvents(RpcUrl, ProgramId)
		time.Sleep(30 * time.Second)
	}
}
