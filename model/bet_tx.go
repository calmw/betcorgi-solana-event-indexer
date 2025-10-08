package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BetTx struct {
	Id                      uint64          `gorm:"primaryKey" json:"id"`
	GameId                  int64           `gorm:"column:game_id;" json:"game_id"`
	OrderId                 int64           `gorm:"column:order_id;" json:"order_id"`
	Player                  string          `gorm:"column:player;type:varchar(100);comment:'player'" json:"player"`
	BridgeMsg               []byte          `gorm:"column:bridge_msg;comment:'跨链nsg'" json:"bridge_msg"`
	Hash                    string          `gorm:"column:hash;unique;comment:'唯一索引'" json:"hash"`
	VoteStatus              int             `gorm:"column:vote_status;default:0;comment:'vote 0失败，1成功'" json:"vote_status"`
	ExecuteStatus           int             `gorm:"column:execute_status;default:0;comment:'execute 0失败，1成功'" json:"execute_status"`
	Amount                  decimal.Decimal `gorm:"column:amount;type:decimal(40,0);comment:'跨链数额'" json:"amount"`
	Fee                     decimal.Decimal `gorm:"column:fee;type:decimal(40,0);comment:'跨链费用'" json:"fee"`
	Caller                  string          `gorm:"column:caller;comment:'链链发起者地址'" json:"caller"`
	Receiver                string          `gorm:"column:receiver;comment:'目标链接受者地址'" json:"receiver"`
	SourceTokenAddress      string          `gorm:"column:source_token_address;comment:'源链token地址'" json:"source_token_address"`
	DestinationTokenAddress string          `gorm:"column:destination_token_address;comment:'目标链token地址'" json:"destination_token_address"`
	BridgeStatus            int             `gorm:"column:bridge_status;type:tinyint;comment:'跨链状态 1 源链deposit成功 2 目标链执行成功 3 失败';default:1" json:"bridge_status"`
	DepositHash             string          `gorm:"column:deposit_hash;comment:'deposit tx hash'" json:"deposit_hash"`
	ExecuteHash             string          `gorm:"column:execute_hash;comment:'execute tx hash'" json:"execute_hash"`
	DepositAt               string          `gorm:"column:deposit_at;comment:'跨链发起时间'" json:"deposit_at"`
	ReceiveAt               string          `gorm:"column:receive_at;comment:'跨链到账时间'" json:"receive_at"`
	DeletedAt               gorm.DeletedAt  `gorm:"index"`
	Version                 uint32          `gorm:"not null;default:0;comment:'乐观锁'" json:"version"`
	BridgeGasFee            decimal.Decimal `gorm:"column:amount;type:decimal(20,0);comment:'跨链桥今日累计Gas费用'" json:"bridge_gas_fee"`
}
