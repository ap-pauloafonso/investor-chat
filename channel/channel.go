package channel

import (
	"context"
	"errors"
	"github.com/ap-pauloafonso/investor-chat/user"
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
	GetChannels(ctx context.Context) ([]string, error)
	SaveChannel(ctx context.Context, name string) error
	GetChannel(ctx context.Context, name string) (string, error)
}

func (s *Service) GetAllChannels(ctx context.Context) ([]string, error) {
	return s.r.GetChannels(ctx)
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

func (s *Service) CreateChannel(ctx context.Context, name string) error {
	if len(name) < 3 {
		return errChannelNameShort
	}
	if len(name) > 100 {
		return errChannelNameLong
	}

	if !isValidChannelName(name) {
		return errInvalidChannelName
	}

	_, err := s.r.GetChannel(ctx, name)
	if err == nil {
		return errChannelExists
	}

	err = s.r.SaveChannel(ctx, name)
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
