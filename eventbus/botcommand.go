package eventbus

import (
	"fmt"
	"github.com/wagslane/go-rabbitmq"
	"time"
)

const commandRequestRoutingKey = "botrequest-command"

const commandResponseRoutingKey = "botresponse-command"

type BotCommandRequest struct {
	Command string
	Channel string
	Time    time.Time
}

type BotCommandResponse struct {
	GeneratedMessage string
	Channel          string
	Time             time.Time
}

func (e *Eventbus) PublishBotCommandRequest(msg string) error {
	err := e.publisher.Publish(
		[]byte(msg),
		[]string{commandRequestRoutingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return fmt.Errorf("error publishing botrequest-command: %w", err)
	}
	return nil
}

func (e *Eventbus) PublishBotCommandResponse(msg string) error {
	err := e.publisher.Publish(
		[]byte(msg),
		[]string{commandResponseRoutingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
	)
	if err != nil {
		return fmt.Errorf("error publishing botresponse-command: %w", err)
	}
	return nil
}

func (e *Eventbus) ConsumeBotCommandRequest(fn func(payload []byte) error) error {
	consumer, err := rabbitmq.NewConsumer(
		e.conn,
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
		return fmt.Errorf("error in ConsumeBotCommandRequest: %w", err)
	}
	e.consumers = append(e.consumers, consumer)
	return nil
}

func (e *Eventbus) ConsumeBotCommandResponse(fn func(payload []byte) error) error {
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
		rabbitmq.WithConsumerOptionsRoutingKey(commandResponseRoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return fmt.Errorf("error in ConsumeBotCommandResponse: %w", err)
	}
	e.consumers = append(e.consumers, consumer)
	return nil
}
