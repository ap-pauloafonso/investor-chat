package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"investorchat/chat"
	"investorchat/queue"
	"investorchat/utils"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	errConnectionNotFound = errors.New("user websocket connection not found")
	errChannelNotFound    = errors.New("channel not found")
)

type ChannelUserConnections struct {
	users        map[string]*websocket.Conn
	sync.RWMutex // for mutual exclusion while operating over users inside a channel
}

func (c *ChannelUserConnections) addUser(user string, conn *websocket.Conn) {
	c.Lock()
	defer c.Unlock()
	c.users[user] = conn
}

func (c *ChannelUserConnections) removeUser(user string) {
	c.Lock()
	defer c.Unlock()
	delete(c.users, user)
}

func (c *ChannelUserConnections) getUser(user string) (*websocket.Conn, bool) {
	c.RLock()
	defer c.RUnlock()
	r, ok := c.users[user]
	return r, ok

}

type ChannelConnections struct {
	channels     map[string]*ChannelUserConnections
	sync.RWMutex // for mutual exclusion while operating over a channel
}

func (c *ChannelConnections) addChannel(channelName string) {
	c.Lock()
	defer c.Unlock()

	c.channels[channelName] = &ChannelUserConnections{
		users: map[string]*websocket.Conn{},
	}
}

func (c *ChannelConnections) addUser(channel, user string, conn *websocket.Conn) {
	slog.Info("trying to add user con map")
	c.RLock()
	channelUsers, ok := c.channels[channel]
	c.RUnlock()
	if !ok {
		c.addChannel(channel)              // first user logged in this channel, create the channel entry on our map
		channelUsers = c.channels[channel] // retry
	}

	channelUsers.addUser(user, conn)
	slog.Info("user added to con map")

}

func (c *ChannelConnections) removeUser(channel, user string) {
	c.RLock()
	defer c.RUnlock()

	users, ok := c.channels[channel]
	if ok {
		users.removeUser(user)
	}

}

func (c *ChannelConnections) getChannelUsers(channel string) (*ChannelUserConnections, bool) {
	c.RLock()
	defer c.RUnlock()
	r, ok := c.channels[channel]
	return r, ok

}

type WebSocketHandler struct {
	channelConnections ChannelConnections
	upgrader           websocket.Upgrader
	q                  *queue.Queue
	onConnectionFn     func(channel, user string)
}

type MessageObj struct {
	Username string
	Channel  string
	Message  string
	Time     time.Time
}

func NewWebSocketHandler(q *queue.Queue) *WebSocketHandler {

	channels := make(map[string]*ChannelUserConnections)

	return &WebSocketHandler{
		channelConnections: ChannelConnections{channels: channels},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		q: q,
	}
}

// OnConnection accepts a function to be fired when the user connects
func (w *WebSocketHandler) OnConnection(fn func(channel, user string)) {
	w.onConnectionFn = fn
}

func (w *WebSocketHandler) HandleRequest(c echo.Context) error {

	// Extract the channel from the route parameter
	channelParam := c.Param("channel")

	if len(channelParam) == 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: "Bad Request"})
	}

	u, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: "Bad Request"})
	}

	slog.Info("[user trying to connection]", "channel", channelParam, "user", u)

	conn, err := w.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	w.channelConnections.addUser(channelParam, u, conn)

	slog.Info("[user connected]", "channel", channelParam, "user", u)

	defer func() {

		defer conn.Close()
		w.channelConnections.removeUser(channelParam, u)
		slog.Info("[user disconnected]", "channel", channelParam, "user", u)

	}()

	go w.onConnectionFn(channelParam, u)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		t := time.Now()

		j, err := json.Marshal(MessageObj{
			u, channelParam, string(p), t,
		})
		if err != nil {
			slog.Error("error serializing MessageObj", err)
			continue
		}

		// send the payload to queue
		err = w.q.PublishUserMessageCommand(string(j))
		if err != nil {
			slog.Error(err.Error())
		}

		// stock bot, if it matches then we push the request to the queue
		if okCheckStockCode, stockCode := checkBot(string(p)); okCheckStockCode {
			stock, _ := json.Marshal(queue.BotCommandRequest{
				Command: stockCode,
				Channel: channelParam,
				Time:    t,
			})
			err := w.q.PublishBotCommandRequest(string(stock))
			if err != nil {
				slog.Error(err.Error())
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

func (w *WebSocketHandler) BroadcastMessage(username, channel, msg string, isBoot bool, t time.Time) error {
	channelUsers, okChannel := w.channelConnections.getChannelUsers(channel)
	if !okChannel {
		return errChannelNotFound
	}

	for _, userC := range channelUsers.users {
		jsonBytes, err := json.Marshal(payload{
			Username: username,
			Msg:      msg,
			IsBot:    isBoot,
			Time:     t,
		})
		if err != nil {
			return err
		}

		err = userC.WriteMessage(websocket.TextMessage, jsonBytes)
		if err != nil {
			slog.Error("error writing to user ws", err)
		}

	}

	return nil

}

func (w *WebSocketHandler) HandleChannelsUpdate(channels []string) error {
	// update server channel connection map
	for _, c := range channels {
		if _, ok := w.channelConnections.getChannelUsers(c); !ok {
			w.channelConnections.addChannel(c)
		}

	}

	// broadcast it to everyone connected in ws
	for _, channeList := range w.channelConnections.channels {
		for _, user := range channeList.users {
			err := user.WriteMessage(websocket.TextMessage, []byte("[channel_list_update]"))
			if err != nil {
				slog.Error("error writing [channel_list_update] to user", err)
			}

		}
	}

	return nil

}

func (w *WebSocketHandler) PrintOnlineUsers() {
	go func() {
		for {
			var args []any
			for k, v := range w.channelConnections.channels {
				args = append(args, k, v.users)
			}
			slog.Info("[online users]", args...)
			time.Sleep(10 * time.Second)
		}
	}()

}

func (w *WebSocketHandler) AddNewChannel(channel string) {
	w.channelConnections.addChannel(channel)
}

func (w *WebSocketHandler) SendRecentMessages(channel, username string, msgs []chat.Message) error {
	channelUsers, okChannel := w.channelConnections.getChannelUsers(channel)
	if !okChannel {
		return errors.New("channel not found")
	}

	userC, ok := channelUsers.getUser(username)
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

	err = userC.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		return fmt.Errorf("ws: error writing recent messages to user : %w", err)
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
