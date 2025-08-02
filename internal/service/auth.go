package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
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

func (s *authService) CreateUser(user *dto.SignUpUser) error {
	userID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("can't generate UUID: %w", err)
	}

	hashedPassword, err := generatePasswordHash(user.Password)
	if err != nil {
		return fmt.Errorf("can't generate password hash: %w", err)
	}

	u := &entities.User{
		ID:       userID,
		Username: user.Username,
		Password: hashedPassword,
		Email:    user.Email,
	}

	return s.repo.CreateUser(u)
}

func (s *authService) GetUserByEmail(email string) (*entities.User, error) {
	return s.repo.GetUserByEmail(email)
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *authService) NewRefreshToken(userID uuid.UUID) (*entities.RefreshToken, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	tokenString := fmt.Sprintf("%x", b)
	expiresAt := time.Now().Add(refreshTTL)

	refreshToken := &entities.RefreshToken{
		UserID:       userID,
		RefreshToken: tokenString,
		ExpiresAt:    expiresAt,
	}

	return refreshToken, s.repo.UpsertRefreshToken(refreshToken)
}

func (s *authService) CheckRefreshToken(userID uuid.UUID, refreshToken string) error {
	return s.repo.CheckRefreshToken(userID, refreshToken)
}

func (s *authService) DeleteRefreshToken(userID uuid.UUID) error {
	return s.repo.DeleteRefreshToken(userID)
}
