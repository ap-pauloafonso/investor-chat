package user

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	r Repository
}

func NewService(userRepository Repository) Service {
	return Service{userRepository}
}

type UserModel struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Repository interface {
	SaveUser(username, password string) error
	GetUser(username string) (*UserModel, error)
}

func (s *Service) Register(username, password string) error {
	if len(username) < 3 {
		return errors.New("username should at least have 3 characteres")
	}

	_, err := s.r.GetUser(username)
	if err == nil {
		return errors.New("username already registered")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return errors.New("problem hashing password")
	}

	err = s.r.SaveUser(username, hashedPassword)
	if err != nil {
		return fmt.Errorf("error storing user: %w", err)
	}
	return nil
}

func (s *Service) Login(username, password string) error {
	user, err := s.r.GetUser(username)
	if err != nil {
		return errors.New("invalid credentials")
	}
	if !checkPasswordHash(password, user.Password) {
		return errors.New("invalid credentials")
	}

	return nil

}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
