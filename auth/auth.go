package auth

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RX90/Chat/auth/pkg"
	"github.com/RX90/Chat/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Server *Server
	DB     *sqlx.DB
}

var refreshTTL = 15 * 24 * time.Hour

func (a *Auth) signUp(c *gin.Context) {
	var input pkg.User

	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("can't bind JSON: %v", err)})
		return
	}

	var exists bool

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(login) = LOWER($1))"
	err := a.DB.QueryRow(query, input.Login).Scan(&exists)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't check login: %v", err)})
		return
	}
	if exists {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"err": "login is already taken"})
		return
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't generate UUID: %v", err)})
		return
	}
	input.ID = userID

	hashedPassword, err := generatePasswordHash(input.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't generate password hash: %v", err)})
		return
	}
	input.Password = hashedPassword

	query = "INSERT INTO users (id, login, password_hash) values ($1, $2, $3)"

	_, err = a.DB.Exec(query, input.ID, input.Login, input.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *Auth) signIn(c *gin.Context) {
	var input pkg.User

	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("can't bind JSON: %v", err)})
		return
	}

	var (
		userID         uuid.UUID
		hashedPassword string
	)

	query := "SELECT id, password_hash FROM users WHERE login = $1"
	err := a.DB.QueryRow(query, input.Login).Scan(&userID, &hashedPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "invalid login or password"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "invalid login or password"})
		return
	}

	accessToken, err := middleware.NewAccessToken(userID.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create access token: %v", err)})
		return
	}

	refreshToken, expiresAt, err := a.newRefreshToken(userID.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  expiresAt,
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (a *Auth) refreshTokens(c *gin.Context) {
	userID := c.GetString("userID")
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil || refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "refresh token is missing"})
		return
	}

	if err := a.checkRefreshToken(userID, refreshToken); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("refresh token is invalid: %v", err)})
		return
	}

	accessToken, err := middleware.NewAccessToken(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create access token: %v", err)})
		return
	}

	refreshToken, expiresAt, err := a.newRefreshToken(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  expiresAt,
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (a *Auth) logout(c *gin.Context) {
	userID := c.GetString("userID")

	if err := a.deleteRefreshToken(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("error occured on deleting refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:   "refreshToken",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *Auth) newRefreshToken(userID string) (string, time.Time, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", time.Time{}, err
	}

	token := fmt.Sprintf("%x", b)
	expiresAt := time.Now().Add(refreshTTL)

	return token, expiresAt, a.upsertRefreshToken(userID, token, expiresAt)
}

func (a *Auth) upsertRefreshToken(userID, refreshToken string, expiresAt time.Time) error {
	tx, err := a.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingTokenId string

	query := `
		SELECT ut.token_id
		FROM users_tokens ut
		INNER JOIN tokens t ON ut.token_id = t.id
		WHERE ut.user_id = $1`
	err = tx.QueryRow(query, userID).Scan(&existingTokenId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// Insert Refresh Token
		query = "INSERT INTO tokens (refresh_token, expires_at) values ($1, $2) RETURNING id"
		row := tx.QueryRow(query, refreshToken, expiresAt)

		var tokenId string

		if err := row.Scan(&tokenId); err != nil {
			return err
		}

		query = "INSERT INTO users_tokens (user_id, token_id) values ($1, $2)"
		_, err = tx.Exec(query, userID, tokenId)
		if err != nil {
			return err
		}
	} else {
		// Update Refresh Token
		query = "UPDATE tokens SET refresh_token = $1, expires_at = $2 WHERE id = $3"
		_, err = tx.Exec(query, refreshToken, expiresAt, existingTokenId)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (a *Auth) checkRefreshToken(userID, refreshToken string) error {
	var (
		tokenID     string
		storedToken string
		expiresAt   time.Time
	)

	query := "SELECT ut.token_id FROM users_tokens ut WHERE ut.user_id = $1"
	err := a.DB.QueryRow(query, userID).Scan(&tokenID)
	if err != nil {
		return err
	}

	query = "SELECT t.refresh_token, t.expires_at FROM tokens t WHERE t.id = $1"
	err = a.DB.QueryRow(query, tokenID).Scan(&storedToken, &expiresAt)
	if err != nil {
		return err
	}

	if storedToken != refreshToken {
		return errors.New("tokens are different")
	}

	if time.Now().After(expiresAt) {
		return errors.New("token has expired")
	}

	return nil
}

func (a *Auth) deleteRefreshToken(userID string) error {
	query := `
		DELETE FROM tokens t
		USING users_tokens ut
		WHERE t.id = ut.token_id AND ut.user_id = $1`

	_, err := a.DB.Exec(query, userID)

	return err
}
