package queue

import (
	"fmt"
	"github.com/wagslane/go-rabbitmq"
)

const messageRoutingKey = "message-command"

func (q *Queue) PublishUserMessageCommand(msg string) error {
	err := q.publisher.Publish(
		[]byte(msg),
		[]string{messageRoutingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return fmt.Errorf("error publishing message-command: %w", err)
	}

	return nil
}

func (q *Queue) ConsumeUserMessageCommandForWSBroadcast(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		q.conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			err := fn(d.Body)
			if err != nil {
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		"",
		rabbitmq.WithConsumerOptionsQueueAutoDelete,
		rabbitmq.WithConsumerOptionsRoutingKey(messageRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	q.consumers = append(q.consumers, consumer)
	return nil
}

func (q *Queue) ConsumeUserMessageCommandForStorage(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		q.conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			err := fn(d.Body)
			if err != nil {
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		"storage-q",
		rabbitmq.WithConsumerOptionsRoutingKey(messageRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	q.consumers = append(q.consumers, consumer)

	return nil
}
