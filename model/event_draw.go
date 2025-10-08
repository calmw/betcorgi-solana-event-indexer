package model

import (
	"betcorgi-event-indexer/db"
	"log"

	"gorm.io/gorm"
)

type EventDraw struct {
	gorm.Model
	Id          uint64 `gorm:"primaryKey" json:"id"`
	GameID      uint64 `gorm:"column:game_id;comment:'game_id'" json:"game_id"`
	OrderId     uint64 `gorm:"column:order_id;comment:'order_id'" json:"order_id"`
	Player      string `gorm:"column:player;type:varchar(100);comment:'player'" json:"player"`
	Amount      string `gorm:"column:amount;type:varchar(100);comment:'amount'" json:"amount"`
	HashExpired bool   `gorm:"column:hash_expired;type:varchar(100);comment:'hash_expired'" json:"hash_expired"`
	Seed        string `gorm:"column:seed;type:varchar(500);comment:'seed'" json:"seed"`
	Signature   string `gorm:"uniqueIndex;column:signature;type:varchar(100);comment:'signature'" json:"signature"`
}

func SaveEventDrawToDB(gameId, orderId uint64, player, amount, seed, sig string, hashExpired bool) error {
	var ev EventDraw
	err := db.DB.Model(&EventDraw{}).Where("signature=? and game_id=?", sig, gameId).First(&ev).Error
	if err == nil {
		log.Println("该记录已经存在，跳过")
		return nil
	}
	return db.DB.Model(&EventDraw{}).Create(&EventDraw{
		GameID:      gameId,
		OrderId:     orderId,
		Player:      player,
		Amount:      amount,
		HashExpired: hashExpired,
		Seed:        seed,
		Signature:   sig,
	}).Error
}
