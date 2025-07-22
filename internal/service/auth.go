package service

import (
	"fmt"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo repo.AuthRepo
}

func newAuthService(r repo.AuthRepo) AuthService {
	return &authService{repo: r}
}

func (s *authService) CreateUser(user domain.User) error {
	userID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("can't generate UUID: %w", err)
	}
	user.ID = userID

	hashedPassword, err := generatePasswordHash(user.Password)
	if err != nil {
		return fmt.Errorf("can't generate password hash: %w", err)
	}
	user.Password = hashedPassword

	return s.repo.CreateUser(user)
}

func (s *authService) GetUser(email string) (*domain.User, error) {
	return s.repo.GetUser(email)
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
