package model

import (
	"betcorgi-event-indexer/db"
	"log"
	"time"

	"gorm.io/gorm"
)

type EventBet struct {
	Id          uint64 `gorm:"primaryKey" json:"id"`
	GameId      uint64 `gorm:"column:game_id;comment:'game_id'" json:"game_id"`
	OrderId     uint64 `gorm:"column:order_id;comment:'order_id'" json:"order_id"`
	Hash        string `gorm:"column:hash;type:varchar(100);comment:'hash'" json:"hash"`
	Data        string `gorm:"column:data;type:varchar(500);comment:'data'" json:"data"`
	Player      string `gorm:"column:player;type:varchar(100);comment:'player'" json:"player"`
	Amount      string `gorm:"column:amount;type:varchar(100);comment:'amount'" json:"amount"`
	ProgramData string `gorm:"column:program_data;type:varchar(500);comment:'program_data'" json:"program_data"`
	Signature   string `gorm:"uniqueIndex;column:signature;type:varchar(100);comment:'signature'" json:"signature"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func SaveEventBetToDB(dataB64 string, gameId, orderId uint64, player, amount, hash, data, sig string) error {
	var ev EventBet
	err := db.DB.Model(&EventBet{}).Where("signature=? and game_id=?", sig, gameId).First(&ev).Error
	if err == nil {
		log.Println("该记录已经存在，跳过")
		return nil
	}
	return db.DB.Model(&EventBet{}).Create(&EventBet{
		GameId:      gameId,
		OrderId:     orderId,
		Player:      player,
		Amount:      amount,
		Hash:        hash,
		Data:        data,
		Signature:   sig,
		ProgramData: dataB64,
	}).Error
}
