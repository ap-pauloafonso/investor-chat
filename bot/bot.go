package bot

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"investorchat/queue"
	"net/http"
	"strings"
)

var errInvalidCsvFormat = errors.New("invalid CSV data format")

type Bot struct {
	q *queue.Queue
}

func NewBot(q *queue.Queue) *Bot {
	return &Bot{q: q}
}

func (b *Bot) Process() error {
	err := b.q.ConsumeBotCommandRequest(func(payload []byte) error {
		var obj queue.BotCommandRequest
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		message, err := GetStockMessage(obj.Command)
		if err != nil {
			return err
		}

		var result queue.BotCommandResponse

		result.GeneratedMessage = message
		result.Channel = obj.Channel
		result.Time = obj.Time

		marshal, err := json.Marshal(result)
		if err != nil {
			return err
		}

		err = b.q.PublishBotCommandResponse(string(marshal))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while processing bot: %w", err)
	}

	return nil
}

func GetStockMessage(stockCode string) (string, error) {
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	reader := csv.NewReader(response.Body)

	_, err = reader.Read() // discard first line
	if err != nil {
		return "", err
	}
	record, err := reader.Read()
	if err != nil {
		return "", err
	}

	if len(record) < 7 {
		return "", errInvalidCsvFormat
	}

	stockName := strings.TrimSpace(record[0])
	stockQuote := strings.TrimSpace(record[6])

	message := fmt.Sprintf("%s quote is $%s per share", stockName, stockQuote)

	if stockQuote == "N/D" {
		return "", errors.New("invalid stock code")
	}

	return message, nil

}
