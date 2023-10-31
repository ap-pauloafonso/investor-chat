package bot

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"net/http"
	"strings"
)

var (
	errInvalidCsvFormat = errors.New("invalid CSV data format")
	errInvalidStockCode = errors.New("invalid stock code")
)

type Bot struct {
	eventbus *eventbus.Eventbus
}

func NewBot(eventbus *eventbus.Eventbus) *Bot {
	return &Bot{eventbus: eventbus}
}

const stockURL = "https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv"

func (b *Bot) Process() error {
	err := b.eventbus.ConsumeBotCommandRequest(func(payload []byte) error {
		var obj eventbus.BotCommandRequest
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		message, err := GetStockMessage(stockURL, obj.Command)
		if err != nil {
			return err
		}

		var result eventbus.BotCommandResponse

		result.GeneratedMessage = message
		result.Channel = obj.Channel
		result.Time = obj.Time

		marshal, err := json.Marshal(result)
		if err != nil {
			return err
		}

		err = b.eventbus.PublishBotCommandResponse(string(marshal))
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

func GetStockMessage(url, stockCode string) (string, error) {
	urlParsed := fmt.Sprintf(url, stockCode)

	response, err := http.Get(urlParsed) //nolint
	if err != nil {
		return "", err
	}

	reader := csv.NewReader(response.Body)

	_, err = reader.Read() // discard first line
	if err != nil {
		return "", errInvalidCsvFormat
	}
	record, err := reader.Read()
	if err != nil {
		return "", errInvalidCsvFormat
	}

	if len(record) < 7 {
		return "", errInvalidCsvFormat
	}

	stockName := strings.TrimSpace(record[0])
	stockQuote := strings.TrimSpace(record[6])

	message := fmt.Sprintf("%s quote is $%s per share", stockName, stockQuote)

	if stockQuote == "N/D" {
		return "", errInvalidStockCode
	}

	return message, nil

}
