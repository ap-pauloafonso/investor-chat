package queue

import (
	"github.com/wagslane/go-rabbitmq"
)

type Queue struct {
	conn      *rabbitmq.Conn
	publisher *rabbitmq.Publisher
	consumers []*rabbitmq.Consumer
}

const exchangeName = "events"
const contentType = "application/json"

func NewQueue() (*Queue, error) {

	rmq, err := rabbitmq.NewConn("amqp://admin:admin@rabbitmq", rabbitmq.WithConnectionOptionsLogging)

	if err != nil {
		return nil, err
	}

	publisher, err := rabbitmq.NewPublisher(
		rmq,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		return nil, err
	}

	return &Queue{
		conn:      rmq,
		publisher: publisher,
	}, nil

}

func (q *Queue) Close() {
	for _, item := range q.consumers {
		item.Close()
	}
	q.publisher.Close()
	q.conn.Close()
}
