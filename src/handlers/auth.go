package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Grizak/isaksweb-auth/src/config"
	"github.com/Grizak/isaksweb-auth/src/store"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Register(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		id := uuid.NewString()
		_, err = store.DB.Exec(
			`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
			id, body.Email, string(hash),
		)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "email": body.Email})
	}
}

func Token(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			GrantType string `json:"grant_type" binding:"required"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if body.GrantType != "password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
			return
		}

		// 1. Look up the user
		var id, hashed string
		err := store.DB.QueryRow(
			`SELECT id, password FROM users WHERE email = ?`, body.Email,
		).Scan(&id, &hashed)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 2. Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(body.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 3. Issue access token (15 min)
		accessToken, err := signJWT(id, body.Email, cfg.JWTSecret, 15*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		// 4. Issue + store refresh token (7 days)
		refreshToken, err := signJWT(id, body.Email, cfg.JWTSecret, 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue refresh token"})
		}
		store.DB.Exec(
			`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
			refreshToken, id, time.Now().Add(7*24*time.Hour),
		)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    15 * time.Minute,
		})
	}
}

// --- Helpers ---

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func signJWT(userID, email, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
