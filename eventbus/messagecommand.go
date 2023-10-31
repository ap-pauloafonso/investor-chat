package eventbus

import (
	"fmt"
	"github.com/wagslane/go-rabbitmq"
)

const messageRoutingKey = "message-command"

func (e *Eventbus) PublishUserMessageCommand(msg string) error {
	err := e.publisher.Publish(
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

func (e *Eventbus) ConsumeUserMessageCommandForWSBroadcast(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		e.conn,
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

	e.consumers = append(e.consumers, consumer)
	return nil
}

func (e *Eventbus) ConsumeUserMessageCommandForStorage(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		e.conn,
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

	e.consumers = append(e.consumers, consumer)

	return nil
}
