package queue

import "github.com/wagslane/go-rabbitmq"

const commandUpdateChannelsRoutingKey = "updatechannels-command"

func (q *Queue) PublishUpdateChannelsCommand() error {
	err := q.publisher.Publish(
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

func (q *Queue) ConsumeUpdateChannelsCommand(fn func(payload []byte) error) error {
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
		rabbitmq.WithConsumerOptionsRoutingKey(commandUpdateChannelsRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	q.consumers = append(q.consumers, consumer)

	return nil
}
