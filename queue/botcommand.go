package queue

import "github.com/wagslane/go-rabbitmq"

const commandRequestRoutingKey = "botrequest-command"

const commandResponseRoutingKey = "botresponse-command"

func (q *Queue) PublishBotCommandRequest(msg string) error {
	err := q.publisher.Publish(
		[]byte(msg),
		[]string{commandRequestRoutingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return err
	}

	return nil
}
func (q *Queue) PublishBotCommandResponse(msg string) error {
	err := q.publisher.Publish(
		[]byte(msg),
		[]string{commandResponseRoutingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue) ConsumeBotCommandRequest(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		q.conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			err := fn(d.Body)
			if err != nil {
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		"botrequest-q",
		rabbitmq.WithConsumerOptionsRoutingKey(commandRequestRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	q.consumers = append(q.consumers, consumer)

	return nil
}

func (q *Queue) ConsumeBotCommandResponse(fn func(payload []byte) error) error {
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
		rabbitmq.WithConsumerOptionsRoutingKey(commandResponseRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	q.consumers = append(q.consumers, consumer)

	return nil
}
