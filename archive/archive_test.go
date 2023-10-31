package archive

import (
	"context"
	"errors"
	"github.com/ap-pauloafonso/investor-chat/user"
	"testing"
	"time"
)

type mockRepository struct {
	recentMessagesErr error
	recentMsgs        map[string][]user.Message
	errToReturn       error
}

func (m *mockRepository) SaveMessage(_ context.Context, c, u, msg string, timestamp time.Time) error {
	if m.recentMsgs == nil {
		m.recentMsgs = map[string][]user.Message{}
	}
	if m.recentMsgs[c] == nil {
		m.recentMsgs[c] = []user.Message{}
	}
	m.recentMsgs[c] = append(m.recentMsgs[c], user.Message{
		Channel:   c,
		User:      u,
		Text:      msg,
		Timestamp: timestamp,
	})
	return m.errToReturn
}

func (m *mockRepository) GetRecentMessages(_ context.Context, channel string, maxMessages int) ([]user.Message, error) {
	if len(m.recentMsgs[channel]) > maxMessages {
		return m.recentMsgs[channel][len(m.recentMsgs[channel])-maxMessages:], m.errToReturn
	}
	return m.recentMsgs[channel], m.recentMessagesErr
}

func TestSaveMessage(t *testing.T) {

	t.Run("Save Message Successfully", func(t *testing.T) {
		repo := &mockRepository{}

		service := NewService(repo, nil)

		timestamp := time.Now()
		err := service.SaveMessage(context.Background(), "channel1", "user1", "Hello", timestamp)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Error Saving Message", func(t *testing.T) {

		repo := &mockRepository{}

		service := NewService(repo, nil)

		repo.errToReturn = errors.New("mock repository error")
		timestamp := time.Now()
		err := service.SaveMessage(context.Background(), "channel1", "user1", "Hello", timestamp)
		if !errors.Is(err, repo.errToReturn) {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
	})
}

func TestGetRecentMessages(t *testing.T) {

	t.Run("Get Recent Messages Successfully", func(t *testing.T) {
		repo := &mockRepository{}

		service := NewService(repo, nil)

		repo.recentMsgs = map[string][]user.Message{
			"channel1": {
				{Channel: "channel1", User: "user1", Text: "Hello", Timestamp: time.Now()},
				{Channel: "channel1", User: "user2", Text: "Hi", Timestamp: time.Now()},
			},
		}

		messages, err := service.GetRecentMessages(context.Background(), "channel1")
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

		service := NewService(repo, nil)

		repo.errToReturn = errors.New("mock repository error")
		messages, err := service.GetRecentMessages(context.Background(), "channel1")
		if !errors.Is(err, errStore) {
			t.Errorf("Expected %v, got %v", repo.errToReturn, err)
		}
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages, got %v", len(messages))
		}
	})
}
