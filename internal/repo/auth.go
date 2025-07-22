package repo

import (
	"github.com/RX90/Chat/internal/domain"
	"gorm.io/gorm"
)

type authRepo struct {
	db *gorm.DB
}

func newAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{db: db}
}

func (r *authRepo) CreateUser(user domain.User) error {
	return r.db.Create(&user).Error
}

func (r *authRepo) GetUser(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}