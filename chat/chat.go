package chat

import (
	"errors"
	"time"
)

type Service struct {
	r Repository
}

func NewService(chatRepository Repository) Service {
	return Service{chatRepository}
}

type Repository interface {
	GetChannels() ([]string, error)
	SaveChannel(name string) error
	GetChannel(name string) (string, error)
	SaveMessage(channel, user, msg string, timestamp time.Time) error
	GetRecentMessages(channel string, maxMessages int) ([]Message, error)
}

type Message struct {
	Channel   string    `json:"channel"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *Service) GetAllChannels() ([]string, error) {
	return s.r.GetChannels()
}

var ChannelExistsErr = errors.New("channel already exists")

func (s *Service) CreateChannel(name string) error {
	if len(name) < 3 {
		return errors.New("channel name needs to have at least 3 characters")
	}

	_, err := s.r.GetChannel(name)
	if err == nil {
		return ChannelExistsErr
	}

	err = s.r.SaveChannel(name)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SaveMessage(channel, user, message string, timestamp time.Time) error {
	return s.r.SaveMessage(channel, user, message, timestamp)
}

func (s *Service) GetRecentMessages(channel string) ([]Message, error) {
	const max = 50
	return s.r.GetRecentMessages(channel, max)
}
