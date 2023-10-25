package archive

import (
	"github.com/ap-pauloafonso/investor-chat/queue"
	"github.com/ap-pauloafonso/investor-chat/user"
	"time"
)

type Repository interface {
	SaveMessage(channel, user, msg string, timestamp time.Time) error
	GetRecentMessages(channel string, maxMessages int) ([]user.Message, error)
}

type Service struct {
	r Repository
	q *queue.Queue
}

func NewService(r Repository, q *queue.Queue) *Service {
	return &Service{r: r, q: q}
}

func (s *Service) SaveMessage(channel, user, message string, timestamp time.Time) error {
	return s.r.SaveMessage(channel, user, message, timestamp)
}

func (s *Service) GetRecentMessages(channel string) ([]user.Message, error) {
	const max = 50
	return s.r.GetRecentMessages(channel, max)
}
