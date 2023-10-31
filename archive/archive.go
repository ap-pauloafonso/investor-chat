package archive

import (
	"context"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/user"
	"time"
)

type Repository interface {
	SaveMessage(ctx context.Context, channel, user, msg string, timestamp time.Time) error
	GetRecentMessages(ctx context.Context, channel string, maxMessages int) ([]user.Message, error)
}

type Service struct {
	r        Repository
	eventbus *eventbus.Eventbus
}

func NewService(r Repository, eventbus *eventbus.Eventbus) *Service {
	return &Service{r: r, eventbus: eventbus}
}

func (s *Service) SaveMessage(ctx context.Context, channel, user, message string, timestamp time.Time) error {
	return s.r.SaveMessage(ctx, channel, user, message, timestamp)
}

func (s *Service) GetRecentMessages(ctx context.Context, channel string) ([]user.Message, error) {
	const max = 50
	return s.r.GetRecentMessages(ctx, channel, max)
}
