package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"investorchat/chat"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type userConnection struct {
	conn *websocket.Conn
}

type listOfUserConnection struct {
	list map[string]userConnection
	mu   sync.Mutex
}

var channelConnections = struct {
	channels map[string]listOfUserConnection
	mu       sync.Mutex
}{
	channels: map[string]listOfUserConnection{},
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (a *App) handleWebSocket(c echo.Context) error {

	// Extract the channel from the route parameter
	channelParam := c.Param("channel")

	if len(channelParam) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: "Bad Request"})
	}

	u, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: "Bad Request"})
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// store connection for later broadcastMessage
	channelConnections.mu.Lock()
	channelUsers, ok := channelConnections.channels[channelParam]
	if !ok {
		panic(err)
	}

	channelUsers.mu.Lock()
	channelUsers.list[u] = userConnection{
		conn: conn,
	}

	channelUsers.mu.Unlock()
	channelConnections.mu.Unlock()

	fmt.Printf("[user connected] user: %s | channel: %s | %v\n", u, channelParam, channelUsers)
	defer func() {
		defer conn.Close()

		channelConnections.mu.Lock()

		connectionList, ok := channelConnections.channels[channelParam]
		if ok {
			delete(connectionList.list, u)
		}

		channelConnections.mu.Unlock()
		fmt.Printf("[user disconneted] user: %s | channel: %s | %v\n", u, channelParam, channelUsers)

	}()

	go func() {
		// send recent messages
		messages, errRecentMessages := a.chatService.GetRecentMessages(channelParam)
		if errRecentMessages != nil {
			fmt.Println("error sending recent messages")
			return
		}

		errRecentMessages = broadcastRecentMessages(channelParam, u, messages)
		if errRecentMessages != nil {
			fmt.Println(err)
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		t := time.Now()

		j, _ := json.Marshal(MessageObj{
			u, channelParam, string(p), t,
		})
		a.q.PublishUserMessageCommand(string(j))

		if okCheckStockCode, stockCode := checkBot(string(p)); okCheckStockCode {
			stock, _ := json.Marshal(BotCommandRequest{
				Command: stockCode,
				Channel: channelParam,
				Time:    t,
			})
			err := a.q.PublishBotCommandRequest(string(stock))
			if err != nil {
				fmt.Println("error sending bot command")
			}
		}

	}
}

type payload struct {
	Username string
	Msg      string
	IsBot    bool
	Time     time.Time
}

func broadcastMessage(username, channel, msg string, isBoot bool, t time.Time) error {
	channelUsers, okChannel := channelConnections.channels[channel]
	if !okChannel {
		return errors.New("channel not found")
	}

	for _, userC := range channelUsers.list {
		jsonBytes, err := json.Marshal(payload{
			Username: username,
			Msg:      msg,
			IsBot:    isBoot,
			Time:     t,
		})
		if err != nil {
			return err
		}

		err = userC.conn.WriteMessage(websocket.TextMessage, jsonBytes)
		if err != nil {
			log.Errorf("error writing to user ws: %s", err.Error())
		}

	}

	return nil

}

func broadcastRecentMessages(channel, username string, msgs []chat.Message) error {
	channelUsers, okChannel := channelConnections.channels[channel]
	if !okChannel {
		return errors.New("channel not found")
	}

	userC, ok := channelUsers.list[username]
	if !ok {
		return errors.New("user connection missing for broadcast recent messages")
	}

	arr := make([]payload, len(msgs))

	for i, m := range msgs {
		arr[i] = payload{
			Username: m.User,
			Msg:      m.Text,
			IsBot:    false,
			Time:     m.Timestamp,
		}
	}

	marshal, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("error encoding array for recent messages: %w", err)
	}

	err = userC.conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		log.Errorf("ws: error writing recent messages to user : %s", err.Error())
	}

	return nil

}

func (a *App) handleChannelsUpdate() error {

	// get all the channels
	channels, err := a.chatService.GetAllChannels()
	if err != nil {
		return err
	}

	// update server channel connection map
	for _, c := range channels {
		if _, ok := channelConnections.channels[c]; !ok {
			channelConnections.mu.Lock()
			channelConnections.channels[c] = listOfUserConnection{
				list: map[string]userConnection{},
			}
			channelConnections.mu.Unlock()
		}

	}

	// broadcast it to everyone connected in ws
	for _, channeList := range channelConnections.channels {
		for _, user := range channeList.list {
			err := user.conn.WriteMessage(websocket.TextMessage, []byte("[channel_list_update]"))
			if err != nil {
				log.Errorf("ws: error writing [channel_list_update] to user : %s", err.Error())
			}

		}
	}

	return nil

}

func checkBot(msg string) (bool, string) {
	regex := regexp.MustCompile(`\/stock=([^\s]+)`)

	matches := regex.FindStringSubmatch(msg)
	if len(matches) > 1 {
		// Extract the stock code from the regex match
		stockCode := matches[1]
		return true, stockCode
	}

	return false, ""
}
