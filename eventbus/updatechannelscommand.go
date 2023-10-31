package eventbus

import (
	"context"
	"github.com/wagslane/go-rabbitmq"
)

const commandUpdateChannelsRoutingKey = "updatechannels-command"

func (e *Eventbus) PublishUpdateChannelsCommand() error {
	err := e.publisher.Publish(
		[]byte("[channel_list_update]"),
		[]string{commandUpdateChannelsRoutingKey},
		rabbitmq.WithPublishOptionsContentType("text/plain"),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return err
	}

	return nil
}

func (e *Eventbus) ConsumeUpdateChannelsCommand(ctx context.Context, fn func(ctx context.Context, payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		e.conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			err := fn(ctx, d.Body)
			if err != nil {
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		"",
		rabbitmq.WithConsumerOptionsQueueAutoDelete,
		rabbitmq.WithConsumerOptionsRoutingKey(commandUpdateChannelsRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	e.consumers = append(e.consumers, consumer)

	return nil
}
