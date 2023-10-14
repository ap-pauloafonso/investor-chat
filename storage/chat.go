package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"investorchat/chat"
	"time"
)

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db}
}

func (r *ChatRepository) GetChannels() ([]string, error) {
	rows, err := r.db.Query(context.Background(), "SELECT name FROM channels")
	if err != nil {
		return nil, fmt.Errorf("error fetching channels: %w", err)
	}
	defer rows.Close()

	channels := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning channels: %w", err)
		}
		channels = append(channels, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over channels: %w", err)
	}

	return channels, nil
}

func (r *ChatRepository) SaveChannel(name string) error {
	_, err := r.db.Exec(context.Background(), "INSERT INTO channels (name) VALUES ($1)", name)
	if err != nil {
		return fmt.Errorf("error saving channel: %w", err)
	}
	return nil
}

func (r *ChatRepository) GetChannel(name string) (string, error) {
	var channelName string
	err := r.db.QueryRow(context.Background(), "SELECT name FROM channels WHERE name = $1", name).Scan(&channelName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("channel with name '%s' not found", name)
		}
		return "", fmt.Errorf("error fetching channel: %w", err)
	}

	return channelName, nil
}

func (r *ChatRepository) SaveMessage(channel, user, msg string, timestamp time.Time) error {
	_, err := r.db.Exec(context.Background(), `
        INSERT INTO messages (channel_name, user_name, message_text, created_at)
        VALUES ($1, $2, $3, $4)`,
		channel, user, msg, timestamp)
	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}
	return nil
}

func (r *ChatRepository) GetRecentMessages(channel string, maxMessages int) ([]chat.Message, error) {
	rows, err := r.db.Query(context.Background(), `
        SELECT channel_name, user_name, message_text, created_at
        FROM (
            SELECT channel_name, user_name, message_text, created_at
            FROM messages
            WHERE channel_name = $1
            ORDER BY created_at DESC
            LIMIT $2
        ) AS recent_messages
        ORDER BY created_at ASC`,
		channel, maxMessages)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages: %w", err)
	}
	defer rows.Close()

	var messages []chat.Message
	for rows.Next() {
		var message chat.Message
		if err := rows.Scan(&message.Channel, &message.User, &message.Text, &message.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}

	return messages, nil
}
