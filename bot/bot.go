package bot

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"investorchat/app"
	"investorchat/queue"
	"net/http"
	"strings"
)

type Bot struct {
	q *queue.Queue
}

func NewBot(q *queue.Queue) *Bot {
	return &Bot{q: q}
}

func (b *Bot) Process() {
	b.q.ConsumeBotCommandRequest(func(payload []byte) error {
		var obj app.BotCommandRequest
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		message, err := GetStockMessage(obj.Command)
		if err != nil {
			return err
		}

		var result app.BotCommandResponse

		result.GeneratedMessage = message
		result.Channel = obj.Channel
		result.Time = obj.Time

		marshal, err := json.Marshal(result)
		if err != nil {
			return err
		}

		b.q.PublishBotCommandResponse(string(marshal))

		return nil
	})
}
func GetStockMessage(stockCode string) (string, error) {
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	reader := csv.NewReader(response.Body)
	reader.Read() // discard first line
	record, err := reader.Read()
	if err != nil {
		return "", err
	}

	if len(record) < 7 {
		return "", fmt.Errorf("Invalid CSV data format")
	}

	stockName := strings.TrimSpace(record[0])
	stockQuote := strings.TrimSpace(record[6])

	message := fmt.Sprintf("%s quote is $%s per share", stockName, stockQuote)

	if stockQuote == "N/D" {
		return "", errors.New("invalid stock code")
	}

	return message, nil

}
