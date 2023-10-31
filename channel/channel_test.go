package channel

import (
	"context"
	"errors"
	"github.com/ap-pauloafonso/investor-chat/user"
	"testing"
	"time"
)

type mockRepository struct {
	getChannelErr     error
	saveChannelErr    error
	recentMessagesErr error
	channels          []string
	channelData       map[string]string
	recentMsgs        map[string][]user.Message
	savedChannel      string
	errToReturn       error
}

func (m *mockRepository) GetChannels(_ context.Context) ([]string, error) {
	return m.channels, m.errToReturn
}

func (m *mockRepository) SaveChannel(_ context.Context, name string) error {
	m.savedChannel = name
	return m.saveChannelErr
}

func (m *mockRepository) GetChannel(_ context.Context, name string) (string, error) {
	return m.channelData[name], m.getChannelErr
}

func (m *mockRepository) SaveMessage(_ context.Context, channel, u, msg string, timestamp time.Time) error {
	if m.recentMsgs == nil {
		m.recentMsgs = map[string][]user.Message{}
	}
	if m.recentMsgs[channel] == nil {
		m.recentMsgs[channel] = []user.Message{}
	}
	m.recentMsgs[channel] = append(m.recentMsgs[channel], user.Message{
		Channel:   channel,
		User:      u,
		Text:      msg,
		Timestamp: timestamp,
	})
	return m.errToReturn
}

func (m *mockRepository) GetRecentMessages(channel string, maxMessages int) ([]user.Message, error) {
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
	msgSent      []user.Message
}

func (m *mockWebSocket) AddNewChannel(channel string) {
	m.addedChannel = channel
}

func (m *mockWebSocket) SendRecentMessages(_, _ string, msgs []user.Message) error {
	m.msgSent = msgs
	return m.sendErr
}

func TestCreateChannel(t *testing.T) {

	t.Run("Valid Channel Creation", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: errChannelExists}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)
		err := service.CreateChannel(context.Background(), "newChannel")
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

		err := service.CreateChannel(context.Background(), "ab")
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
		err := service.CreateChannel(context.Background(), longName)
		if err != errChannelNameLong {
			t.Errorf("Expected %v, got %v", errChannelNameLong, err)
		}
	})

	t.Run("Invalid Channel Name", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel(context.Background(), "invalid_name!")
		if err != errInvalidChannelName {
			t.Errorf("Expected %v, got %v", errInvalidChannelName, err)
		}
	})

	t.Run("Channel Already Exists", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: nil}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel(context.Background(), "channel1")
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

		err := service.CreateChannel(context.Background(), "newChannel")
		if err != saveErr {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
	})

	t.Run("Error Publishing Update Channels Command", func(t *testing.T) {
		repo := &mockRepository{channels: []string{"channel1"}, getChannelErr: errChannelExists, saveChannelErr: nil, errToReturn: errors.New("mock queue error")}
		queue := &mockQueue{}
		ws := &mockWebSocket{}
		service := NewService(repo, queue, ws)

		err := service.CreateChannel(context.Background(), "newChannel")
		if err != queue.errToReturn {
			t.Errorf("Expected %v, got %v", queue.errToReturn, err)
		}
	})
}
