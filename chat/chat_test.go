package chat

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type mockRepository struct {
	getChannelErr     error
	saveChannelErr    error
	recentMessagesErr error
	channels          []string
	channelData       map[string]string
	recentMsgs        map[string][]Message
	savedChannel      string
	errToReturn       error
}

func (m *mockRepository) GetChannels() ([]string, error) {
	return m.channels, m.errToReturn
}

func (m *mockRepository) SaveChannel(name string) error {
	m.savedChannel = name
	return m.saveChannelErr
}

func (m *mockRepository) GetChannel(name string) (string, error) {
	return m.channelData[name], m.getChannelErr
}

func (m *mockRepository) SaveMessage(channel, user, msg string, timestamp time.Time) error {
	if m.recentMsgs == nil {
		m.recentMsgs = map[string][]Message{}
	}
	if m.recentMsgs[channel] == nil {
		m.recentMsgs[channel] = []Message{}
	}
	m.recentMsgs[channel] = append(m.recentMsgs[channel], Message{
		Channel:   channel,
		User:      user,
		Text:      msg,
		Timestamp: timestamp,
	})
	return m.errToReturn
}

func (m *mockRepository) GetRecentMessages(channel string, maxMessages int) ([]Message, error) {
	if len(m.recentMsgs[channel]) > maxMessages {
		return m.recentMsgs[channel][len(m.recentMsgs[channel])-maxMessages:], m.errToReturn
	}
	return m.recentMsgs[channel], m.recentMessagesErr
}

type mockQueue struct {
	errToReturn error
}

func (m *mockQueue) PublishUpdateChannelsCommand() error {
	return m.errToReturn
}

type mockWebSocket struct {
	addedChannel string
	sendErr      error
	msgSent      []Message
}

func (m *mockWebSocket) AddNewChannel(channel string) {
	m.addedChannel = channel
}

func (m *mockWebSocket) SendRecentMessages(channel, user string, msgs []Message) error {
	m.msgSent = msgs
	return m.sendErr
}

func TestCreateChannel(t *testing.T) {

	t.Run("Valid Channel Creation", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: errChannelExists}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)
		err := service.CreateChannel("newChannel")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if repo.savedChannel != "newChannel" {
			t.Errorf("Expected channel 'newChannel' to be saved, got %v", repo.savedChannel)
		}
		if ws.addedChannel != "newChannel" {
			t.Errorf("Expected channel 'newChannel' to be added to WebSocket, got %v", ws.addedChannel)
		}
	})

	t.Run("Short Channel Name", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel("ab")
		if err != errChannelNameShort {
			t.Errorf("Expected %v, got %v", errChannelNameShort, err)
		}
	})

	t.Run("Long Channel Name", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		longName := "looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo"
		err := service.CreateChannel(longName)
		if err != errChannelNameLong {
			t.Errorf("Expected %v, got %v", errChannelNameLong, err)
		}
	})

	t.Run("Invalid Channel Name", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel("invalid_name!")
		if err != errInvalidChannelName {
			t.Errorf("Expected %v, got %v", errInvalidChannelName, err)
		}
	})

	t.Run("Channel Already Exists", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: nil}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel("channel1")
		if err != errChannelExists {
			t.Errorf("Expected %v, got %v", errChannelExists, err)
		}
	})

	t.Run("Error Saving Channel", func(t *testing.T) {
		saveErr := errors.New("save err")
		repo := &mockRepository{channels: []string{"channel1"}, saveChannelErr: saveErr, getChannelErr: errChannelExists}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel("newChannel")
		if err != saveErr {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
	})

	t.Run("Error Publishing Update Channels Command", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: errChannelExists, saveChannelErr: nil, errToReturn: errors.New("mock queue error")}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel("newChannel")
		if err != queue.errToReturn {
			t.Errorf("Expected %v, got %v", queue.errToReturn, err)
		}
	})
}

func TestUserConnected(t *testing.T) {

	t.Run("User Connected Successfully", func(t *testing.T) {
		repo := &mockRepository{}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		repo.recentMsgs = map[string][]Message{
			"channel1": {
				{Channel: "channel1", User: "user1", Text: "Hello", Timestamp: time.Now()},
				{Channel: "channel1", User: "user2", Text: "Hi", Timestamp: time.Now()},
			},
		}

		service.UserConnected("channel1", "user1")
		if !reflect.DeepEqual(ws.msgSent, repo.recentMsgs["channel1"]) {
			t.Errorf("Expected 'channel1' to be added to WebSocket, got %v", ws.addedChannel)
		}
	})

	t.Run("Error getting Recent Messages", func(t *testing.T) {

		repo := &mockRepository{recentMessagesErr: errors.New("failed to get list of")}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		repo.recentMsgs = map[string][]Message{
			"channel1": {
				{Channel: "channel1", User: "user1", Text: "Hello", Timestamp: time.Now()},
				{Channel: "channel1", User: "user2", Text: "Hi", Timestamp: time.Now()},
			},
		}

		ws.sendErr = errors.New("mock websocket error")

		service.UserConnected("channel1", "user1")
		if ws.msgSent != nil {
			t.Errorf("expected msgSent to be nil, since no msg were sent")
		}
	})

}

func TestSaveMessage(t *testing.T) {

	t.Run("Save Message Successfully", func(t *testing.T) {
		repo := &mockRepository{}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		timestamp := time.Now()
		err := service.SaveMessage("channel1", "user1", "Hello", timestamp)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Error Saving Message", func(t *testing.T) {

		repo := &mockRepository{}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		repo.errToReturn = errors.New("mock repository error")
		timestamp := time.Now()
		err := service.SaveMessage("channel1", "user1", "Hello", timestamp)
		if err != repo.errToReturn {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
	})
}

func TestGetRecentMessages(t *testing.T) {

	t.Run("Get Recent Messages Successfully", func(t *testing.T) {
		repo := &mockRepository{}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		repo.recentMsgs = map[string][]Message{
			"channel1": {
				{Channel: "channel1", User: "user1", Text: "Hello", Timestamp: time.Now()},
				{Channel: "channel1", User: "user2", Text: "Hi", Timestamp: time.Now()},
			},
		}

		messages, err := service.GetRecentMessages("channel1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("Expected 2 messages, got %v", len(messages))
		}
	})

	t.Run("Error Getting Recent Messages", func(t *testing.T) {
		errStore := errors.New("some error")
		repo := &mockRepository{recentMessagesErr: errStore}
		queue := &mockQueue{}
		ws := &mockWebSocket{}

		service := NewService(repo, queue, ws)

		repo.errToReturn = errors.New("mock repository error")
		messages, err := service.GetRecentMessages("channel1")
		if err != errStore {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages, got %v", len(messages))
		}
	})
}
