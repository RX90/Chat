package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo repo.AuthRepo
}

var (
	refreshTTL = 15 * 24 * time.Hour
)

func newAuthService(r repo.AuthRepo) AuthService {
	return &authService{repo: r}
}

func (s *authService) CreateUser(user *domain.User) error {
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

func (s *authService) GetUserByEmail(email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(email)
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *authService) NewRefreshToken(userID uuid.UUID) (*domain.Token, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	tokenString := fmt.Sprintf("%x", b)
	expiresAt := time.Now().Add(refreshTTL)

	refreshToken := &domain.Token{
		UserID: userID,
		RefreshToken: tokenString,
		ExpiresAt: expiresAt,
	}

	return refreshToken, s.repo.UpsertRefreshToken(refreshToken)
}
