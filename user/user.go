package user

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	errInvalidCredentials    = errors.New("invalid credentials")
	errUserNamePasswordShort = errors.New("invalid user name or password: needs to have at least 3 characters")
	errUserNamePasswordLong  = errors.New("invalid user name or password: exceed the max amount of 50 characters")
	errUsernameAlreadyTaken  = errors.New("username already registered")
	errHashingPassword       = errors.New("problem hashing password")
	errStoringUser           = errors.New("error storing user")
)

type Service struct {
	r Repository
}

type Message struct {
	Channel   string    `json:"channel"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func NewService(userRepository Repository) *Service {
	return &Service{userRepository}
}

type Model struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Repository interface {
	SaveUser(username, password string) error
	GetUser(username string) (*Model, error)
}

func (s *Service) Register(username, password string) error {
	if len(username) < 3 || len(password) < 3 {
		return errUserNamePasswordShort
	}

	if len(username) > 50 || len(password) > 50 {
		return errUserNamePasswordLong
	}
	_, err := s.r.GetUser(username)
	if err == nil {
		return errUsernameAlreadyTaken
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return errHashingPassword
	}

	err = s.r.SaveUser(username, hashedPassword)
	if err != nil {
		return errStoringUser
	}

	return nil
}

func (s *Service) Login(username, password string) error {
	user, err := s.r.GetUser(username)
	if err != nil {
		return errInvalidCredentials
	}

	if !checkPasswordHash(password, user.Password) {
		return errInvalidCredentials
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
