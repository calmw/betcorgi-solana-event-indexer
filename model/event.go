package model

import (
	"betcorgi-event-indexer/db"
	"log"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Id            uint64 `gorm:"primaryKey" json:"id"`
	Signature     string `gorm:"uniqueIndex;column:signature;type:varchar(100);comment:'signature'" json:"signature"`
	Slot          uint64 `gorm:"column:slot;type:decimal(30,0);comment:'slot'" json:"slot"`
	Discriminator string `gorm:"comment:'discriminator'" json:"discriminator"`
	Payload       []byte `gorm:"type:longblob;comment:'payload'" json:"payload"`
}

func SaveEventToDB(sig string, slot uint64, discriminator string, payload []byte) error {
	var ev Event
	err := db.DB.Model(&Event{}).Where("signature=? and discriminator=?", sig, discriminator).First(&ev).Error
	if err == nil {
		log.Println("该记录已经存在，跳过")
		return nil
	}
	return db.DB.Model(&Event{}).Create(&Event{
		Signature:     sig,
		Slot:          slot,
		Discriminator: discriminator,
		Payload:       payload,
	}).Error
}

func GetLastProcessedSlot() uint64 {
	var record Event
	if err := db.DB.Model(&Event{}).Order("slot desc").First(&record).Error; err != nil {
		return 0
	}
	return record.Slot
}
