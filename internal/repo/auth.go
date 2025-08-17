package repo

import (
	"errors"
	"time"

	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type authRepo struct {
	db *gorm.DB
}

func newAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{db: db}
}

func (r *authRepo) CreateUser(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *authRepo) GetUserByID(userID uuid.UUID) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepo) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepo) UpsertRefreshToken(token *entities.RefreshToken) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"refresh_token", "expires_at"}),
	}).Create(token).Error
}

func (r *authRepo) CheckRefreshToken(userID uuid.UUID, refreshToken string) error {
	var token entities.RefreshToken

	err := r.db.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("refresh token not found")
		}
		return err
	}

	if time.Now().After(token.ExpiresAt) {
		return errors.New("refresh token has expired")
	}

	return nil
}

func (r *authRepo) DeleteRefreshToken(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&entities.RefreshToken{}).Error
}
