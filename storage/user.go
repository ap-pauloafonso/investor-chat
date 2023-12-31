package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/user"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) SaveUser(ctx context.Context, username, password string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
	if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUser(ctx context.Context, username string) (*user.Model, error) {
	var storedPassword string
	err := r.db.QueryRow(ctx, "SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return &user.Model{
		Username: username,
		Password: storedPassword,
	}, nil
}
