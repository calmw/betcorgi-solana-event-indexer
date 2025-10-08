package main

import (
	"betcorgi-event-indexer/cmd/server/service"
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"log"
	"os"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_gin"

	"github.com/gin-gonic/gin"
)

func main() {

	db.InitMysql()

	//自动迁移
	err := db.DB.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").AutoMigrate(&model.Event{}, &model.EventBet{}, &model.EventDraw{})
	if err != nil {
		log.Println("db AutoMigrate err: ", err)
	}

	router := gin.Default()
	// 创建限速器,每秒5次
	limiter := tollbooth.NewLimiter(5, nil)
	// 使用限速中间件
	router.GET("/bet", tollbooth_gin.LimitHandler(limiter), service.EventBet)
	router.GET("/draw", tollbooth_gin.LimitHandler(limiter), service.EventDraw)
	router.GET("/healthy", tollbooth_gin.LimitHandler(limiter), service.Healthy)
	addr := os.Getenv("LISTEN_ADDR")
	if len(addr) == 0 {
		addr = "0.0.0.0:8080"
	}
	_ = router.Run(addr)
}
