package service

import (
	"betcorgi-event-indexer/db"
	"betcorgi-event-indexer/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func EventDraw(c *gin.Context) {
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
	var records []model.EventDraw
	tx := db.DB.Model(&model.EventDraw{}).Order("created_at desc")

	if len(q.Signature) > 0 {
		tx = tx.Where("signature=?", strings.ToLower(q.Signature))
	}
	if q.GameId > 0 {
		tx = tx.Where("game_id=?", q.GameId)
	}
	if len(q.Player) > 0 {
		tx = tx.Where("player=?", q.Player)
	}
	if q.OrderId > 0 {
		tx = tx.Where("order_id=?", q.OrderId)
	}
	tx.Count(&total)
	err := tx.Offset(offset).Order("order_id DESC, created_at DESC").Limit(pageSize).Find(&records).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "系统繁忙",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":  0,
		"msg":   "OK",
		"total": total,
		"data":  records,
	})
}
