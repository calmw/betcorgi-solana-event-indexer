package fetch

import (
	"betcorgi-event-indexer/model"
	event2 "betcorgi-event-indexer/monitor/event"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func FetchMissingEvents(rpcURL, programID string) {
	log.Println("补漏事件开始")
	defer func() {
		log.Println("补漏事件结束")
	}()
	lastSlot := model.GetLastProcessedSlot()
	if lastSlot == 0 {
		return
	}

	// 第一步：获取签名列表
	sigParams := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSignaturesForAddress",
		"params": []interface{}{
			programID,
			map[string]interface{}{
				"limit":      100,
				"commitment": "finalized",
			},
		},
	}

	sigBytes, err := jsonPost(rpcURL, sigParams)
	if err != nil {
		log.Println("⚠️ getSignaturesForAddress 失败:", err)
		return
	}

	var sigResp struct {
		Result []struct {
			Signature string `json:"signature"`
			Slot      uint64 `json:"slot"`
		} `json:"result"`
	}

	if err := json.Unmarshal(sigBytes, &sigResp); err != nil {
		log.Println("⚠️ 签名列表解析失败:", err)
		return
	}

	// 第二步：依次获取交易详情
	for _, tx := range sigResp.Result {
		if tx.Slot <= lastSlot {
			continue
		}

		txParams := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getTransaction",
			"params": []interface{}{
				tx.Signature,
				map[string]interface{}{
					"encoding":   "json",
					"commitment": "finalized",
				},
			},
		}

		txBytes, err := jsonPost(rpcURL, txParams)
		if err != nil {
			log.Println("⚠️ getTransaction 失败:", err)
			continue
		}

		var txResp struct {
			Result struct {
				Slot        uint64 `json:"slot"`
				Transaction struct {
					Signatures []string `json:"signatures"`
				} `json:"transaction"`
				Meta struct {
					LogMessages []string `json:"logMessages"`
				} `json:"meta"`
			} `json:"result"`
		}

		if err := json.Unmarshal(txBytes, &txResp); err != nil {
			log.Println("⚠️ JSON解析失败:", err)
			continue
		}

		signature := tx.Signature
		slot := txResp.Result.Slot

		for _, logLine := range txResp.Result.Meta.LogMessages {
			if strings.HasPrefix(logLine, "Program data: ") {
				dataB64 := strings.TrimPrefix(logLine, "Program data: ")
				event2.HandleProgramData(dataB64, signature, slot)
			}
		}
	}
}

// jsonPost 简单 POST JSON helper
func jsonPost(url string, data interface{}) ([]byte, error) {
	body, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	return io.ReadAll(resp.Body)
}
