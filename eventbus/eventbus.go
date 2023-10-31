package eventbus

import (
	"github.com/wagslane/go-rabbitmq"
)

type Eventbus struct {
	conn      *rabbitmq.Conn
	publisher *rabbitmq.Publisher
	consumers []*rabbitmq.Consumer
}

const exchangeName = "events"
const contentType = "application/json"

func New(rabbitmqURL string) (*Eventbus, error) {

	rmq, err := rabbitmq.NewConn(rabbitmqURL, rabbitmq.WithConnectionOptionsLogging)

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

	return &Eventbus{
		conn:      rmq,
		publisher: publisher,
	}, nil

}

func (e *Eventbus) Close() {
	for _, item := range e.consumers {
		item.Close()
	}
	e.publisher.Close()
	e.conn.Close()
}
