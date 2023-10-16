package user

import (
	"errors"
	"testing"
)

type mockRepository struct {
	userData    map[string]*Model
	saveUserErr error
	getUserErr  error
}

func (m *mockRepository) SaveUser(username, password string) error {
	if m.saveUserErr != nil {
		return m.saveUserErr
	}

	if _, exists := m.userData[username]; exists {
		return errUsernameAlreadyTaken
	}

	m.userData[username] = &Model{Username: username, Password: password}
	return nil
}

func (m *mockRepository) GetUser(username string) (*Model, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}

	user, exists := m.userData[username]
	if !exists {
		return nil, errors.New("not found")
	}
	return user, nil
}
func TestRegister(t *testing.T) {
	repo := &mockRepository{userData: make(map[string]*Model)}
	service := NewService(repo)

	t.Run("Valid Registration", func(t *testing.T) {
		err := service.Register("user1", "password1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Short Username/Password", func(t *testing.T) {
		err := service.Register("u", "p")
		if err != errUserNamePasswordShort {
			t.Errorf("Expected %v, got %v", errUserNamePasswordShort, err)
		}
	})

	t.Run("Long Username/Password", func(t *testing.T) {
		longStr := "loooooooooooooooooooooooooooooooooooooooooooooooooooooong"
		err := service.Register(longStr, longStr)
		if err != errUserNamePasswordLong {
			t.Errorf("Expected %v, got %v", errUserNamePasswordLong, err)
		}
	})

	t.Run("Username Already Taken", func(t *testing.T) {
		service.Register("user2", "password2")        // Register a user
		err := service.Register("user2", "password2") // Try to register the same user again
		if err != errUsernameAlreadyTaken {
			t.Errorf("Expected %v, got %v", errUsernameAlreadyTaken, err)
		}
	})

	t.Run("Error Storing User", func(t *testing.T) {
		// Simulate an error when storing the user
		repo.saveUserErr = errStoringUser // Reset the saveUserErr
		repo.getUserErr = nil
		err := service.Register("user4", "password4")
		if err != errStoringUser {
			t.Errorf("Expected %v, got %v", errStoringUser, err)
		}
	})
}

func TestLogin(t *testing.T) {
	repo := &mockRepository{userData: make(map[string]*Model)}
	service := NewService(repo)
	service.Register("user5", "password5")

	t.Run("Valid Login", func(t *testing.T) {
		err := service.Login("user5", "password5")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		err := service.Login("user5", "wrongPassword")
		if err != errInvalidCredentials {
			t.Errorf("Expected %v, got %v", errInvalidCredentials, err)
		}
	})
}

func Test_hashPassword(t *testing.T) {

	pass := "123"

	v1, _ := hashPassword(pass)

	v2, _ := hashPassword(pass)

	if !checkPasswordHash(pass, v1) {
		t.Errorf("aaaaaaaaaa")
	}

	if !checkPasswordHash(pass, v2) {
		t.Errorf("bbbbbbbbbbb")
	}

}
