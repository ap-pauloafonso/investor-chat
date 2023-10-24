package channel

import (
	"errors"
	"investorchat/user"
	"regexp"
)

var (
	errChannelExists      = errors.New("channel already exists")
	errChannelNameShort   = errors.New("invalid channel name: needs to have at least 3 characters")
	errChannelNameLong    = errors.New("invalid channel name: exceed the max amount of 100 characters")
	errInvalidChannelName = errors.New("invalid channel name: only letters and numbers are allowed")
)

type Service struct {
	r         Repository
	q         Queue
	websocket WebSocket
}

func NewService(chatRepository Repository, q Queue, w WebSocket) *Service {
	return &Service{chatRepository, q, w}
}

type Repository interface {
	GetChannels() ([]string, error)
	SaveChannel(name string) error
	GetChannel(name string) (string, error)
}

type ArchiveServer interface {
	GetRecentMessages(channel string) ([]user.Message, error)
}

func (s *Service) GetAllChannels() ([]string, error) {
	return s.r.GetChannels()
}

type Queue interface {
	PublishUpdateChannelsCommand() error
}

type WebSocket interface {
	AddNewChannel(channel string)
	SendRecentMessages(channel, user string, msgs []user.Message) error
}

var validChannelRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")

func isValidChannelName(name string) bool {
	return validChannelRegex.MatchString(name)
}

func (s *Service) CreateChannel(name string) error {
	if len(name) < 3 {
		return errChannelNameShort
	}
	if len(name) > 100 {
		return errChannelNameLong
	}

	if !isValidChannelName(name) {
		return errInvalidChannelName
	}

	_, err := s.r.GetChannel(name)
	if err == nil {
		return errChannelExists
	}

	err = s.r.SaveChannel(name)
	if err != nil {
		return err
	}

	err = s.q.PublishUpdateChannelsCommand()
	if err != nil {
		return err
	}

	s.websocket.AddNewChannel(name)

	return nil
}
