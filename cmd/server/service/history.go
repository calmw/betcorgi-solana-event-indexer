package service

import (
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Query struct {
	PageNum           int    `form:"page_num"`
	PageSize          int    `form:"page_size"`
	From              string `form:"from"`
	Recipient         string `form:"recipient"`
	FromToken         string `form:"from_token"`
	ToToken           string `form:"to_token"`
	Status            int    `form:"status"`
	DepositAtStart    int64  `form:"deposit_at_start"`
	DepositAtEnd      int64  `form:"deposit_at_end"`
	SourceTxHash      string `form:"source_tx_hash"`
	DestinationTxHash string `form:"destination_tx_hash"`
}

type Response struct {
	Id                      uint64          `json:"id"`
	ResourceId              string          `json:"resource_id"`
	ReceiveAmount           decimal.Decimal `json:"receive_amount"`
	Amount                  decimal.Decimal `json:"amount"`
	From                    string          `json:"from"`
	Recipient               string          `json:"recipient"`
	SourceChain             string          `json:"source_chain"`
	SourceToken             string          `json:"source_token"`
	SourceTokenAddress      string          `json:"source_token_address"`
	SourceTxHash            string          `json:"source_tx_hash"`
	DestinationChain        string          `json:"destination_chain"`
	DestinationToken        string          `json:"destination_token"`
	DestinationTokenAddress string          `json:"destination_token_address"`
	DestinationTxHash       string          `json:"destination_tx_hash"`
	SourceStatus            int             `json:"source_status"`
	DestinationStatus       int             `json:"destination_status"`
	BridgeStatus            int             `json:"bridge_status"`
	DepositAt               string          `json:"deposit_at"`
	ReceiveAt               string          `json:"receive_at"`
	Elapse                  int64           `json:"elapse"`
}

func BridgeTx(c *gin.Context) {
	var q Query
	if c.ShouldBindQuery(&q) != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "参数错误",
		})
		return
	}

	var total int64
	var page = 1
	var pageSize = 10
	if q.PageNum > 0 {
		page = q.PageNum
	}
	if q.PageSize > 0 {
		pageSize = q.PageSize
	}
	offset := (page - 1) * pageSize

	// 更新记录
	var records []model.BridgeTx
	var data []Response
	tx := db.DB.Model(&model.BridgeTx{}).Order("deposit_at desc")

	if len(q.From) > 0 {
		tx = tx.Where("caller=?", strings.ToLower(q.From))
	}
	if len(q.Recipient) > 0 {
		tx = tx.Where("receiver=?", strings.ToLower(q.Recipient))
	}
	if len(q.FromToken) > 0 {
		tx = tx.Where("source_token_address=?", strings.ToLower(q.FromToken))
	}
	if len(q.ToToken) > 0 {
		tx = tx.Where("destination_token_address=?", strings.ToLower(q.ToToken))
	}
	if q.Status > 0 {
		tx = tx.Where("bridge_status=?", q.Status)
	}
	if q.DepositAtStart > 0 {
		tx = tx.Where("deposit_at >=?", time.Unix(q.DepositAtStart, 0))
	}
	if q.DepositAtEnd > 0 {
		tx = tx.Where("deposit_at <=?", time.Unix(q.DepositAtEnd, 0))
	}
	if len(q.SourceTxHash) > 0 {
		sourceTxHash := ""
		sourceTxHashWith0x := ""
		q.SourceTxHash = strings.ToLower(q.SourceTxHash)
		if !strings.HasPrefix(q.SourceTxHash, "0x") {
			sourceTxHash = q.SourceTxHash
			sourceTxHashWith0x = "0x" + q.SourceTxHash
		}
		if strings.HasPrefix(q.SourceTxHash, "0x") {
			sourceTxHashWith0x = q.SourceTxHash
			sourceTxHash = strings.TrimPrefix(q.SourceTxHash, "0x")
		}
		tx = tx.Where("deposit_hash=? or deposit_hash=?", sourceTxHash, sourceTxHashWith0x)
	}
	if len(q.DestinationTxHash) > 0 {
		q.DestinationTxHash = strings.ToLower(q.DestinationTxHash)
		if !strings.HasPrefix(q.DestinationTxHash, "0x") {
			q.DestinationTxHash = "0x" + q.DestinationTxHash
		}
		tx = tx.Where("execute_hash=?", q.DestinationTxHash)
	}
	tx.Count(&total)
	err := tx.Offset(offset).Order("deposit_at DESC").Limit(pageSize).Find(&records).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "系统繁忙",
		})
		return
	}

	for _, record := range records {
		sourceToken := ""
		info, err := model.GetTokenInfo(record.SourceChainId, record.SourceTokenAddress)
		if err == nil {
			sourceToken = info.TokenName
		}
		destinationToken := ""
		info, err = model.GetTokenInfo(record.DestinationChainId, record.DestinationTokenAddress)
		if err == nil {
			destinationToken = info.TokenName
		}
		var receiveAt int64
		var depositAt int64
		depositAtStr := ""
		receiveAtStr := ""
		if len(record.DepositAt) > 0 {
			depositAt, _ = utils.DatetimeToUnix(record.DepositAt)
			depositAtStr = utils.TimestampToDatetime(depositAt)
		}
		if len(record.ReceiveAt) > 0 {
			receiveAt, _ = utils.DatetimeToUnix(record.ReceiveAt)
			receiveAtStr = utils.TimestampToDatetime(receiveAt)
		}

		// source_status int 跨链状态          1 源链pending    	2 源链success   				3 源链failed
		// destination_status int 跨链状态     1 目标链pending  	2 目标链success  			3 目标链failed
		// status int 跨链状态                 1 源链pending    	2 源链success/目标链pending  	3 目标链success
		sourceStatus := 1
		destinationStatus := 1
		bridgeStatus := 1
		var elapse int64
		switch record.VoteStatus {
		case 0:
			switch record.ExecuteStatus {
			case 0:
				sourceStatus = 1
				destinationStatus = 1
				bridgeStatus = 1
			case 1:
				sourceStatus = 2
				destinationStatus = 1
				bridgeStatus = 2
			}
		case 1:
			switch record.ExecuteStatus {
			case 0:
				sourceStatus = 2
				destinationStatus = 1
				bridgeStatus = 2

			case 1:
				sourceStatus = 2
				destinationStatus = 2
				bridgeStatus = 3
				elapse = receiveAt - depositAt
			}
		}

		data = append(data, Response{
			Id:                      record.Id,
			ResourceId:              string(record.ResourceId),
			Amount:                  record.Amount,
			ReceiveAmount:           record.Amount.Sub(record.Fee),
			From:                    record.Caller,
			Recipient:               record.Receiver,
			SourceChain:             record.SourceChainId.String(),
			SourceToken:             sourceToken,
			SourceTokenAddress:      record.SourceTokenAddress,
			SourceTxHash:            record.DepositHash,
			DestinationChain:        record.DestinationChainId.String(),
			DestinationToken:        destinationToken,
			DestinationTokenAddress: record.DestinationTokenAddress,
			DestinationTxHash:       record.ExecuteHash,
			BridgeStatus:            bridgeStatus,
			SourceStatus:            sourceStatus,
			DestinationStatus:       destinationStatus,
			//DepositAt:               utils.TimestampToDatetime(depositAt),
			//ReceiveAt:               utils.TimestampToDatetime(receiveAt),
			DepositAt: depositAtStr,
			ReceiveAt: receiveAtStr,
			Elapse:    elapse,
		})
	}

	c.JSON(200, gin.H{
		"code":  0,
		"msg":   "OK",
		"total": total,
		"data":  data,
	})
}
