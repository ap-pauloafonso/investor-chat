package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/user"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db}
}

func (m *MessageRepository) SaveMessage(channel, user, msg string, timestamp time.Time) error {
	_, err := m.db.Exec(context.Background(), `
        INSERT INTO messages (channel_name, user_name, message_text, created_at)
        VALUES ($1, $2, $3, $4)`,
		channel, user, msg, timestamp)
	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}
	return nil
}

func (m *MessageRepository) GetRecentMessages(channel string, maxMessages int) ([]user.Message, error) {
	rows, err := m.db.Query(context.Background(), `
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

	var messages []user.Message
	for rows.Next() {
		var message user.Message
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
