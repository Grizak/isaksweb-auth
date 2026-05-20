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

// POST /register
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

// POST /oauth/token
func Token(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			GrantType    string `json:"grant_type" binding:"required"`
			Email        string `json:"email"`
			Password     string `json:"password"`
			Code         string `json:"code"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if body.GrantType == "password" {
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
				"expires_in":    900,
			})
		} else if body.GrantType == "authorization_code" {
			// Validate client credentials
			var storedSecret, storedRedirect string
			err := store.DB.QueryRow(
				`SELECT secret, redirect_uri FROM clients WHERE id = ?`, body.ClientID,
			).Scan(&storedSecret, &storedRedirect)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown client"})
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(storedSecret), []byte(body.ClientSecret)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client secret"})
				return
			}
			if storedRedirect != body.RedirectURI {
				c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri mismatch"})
				return
			}

			// Validate the code
			var userID string
			var expiresAt time.Time
			var used int
			err = store.DB.QueryRow(
				`SELECT user_id, expires_at, used FROM authorization_codes WHERE code = ? AND client_id = ?`,
				body.Code, body.ClientID,
			).Scan(&userID, &expiresAt, &used)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
				return
			}
			if used == 1 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "code already used"})
				return
			}
			if time.Now().After(expiresAt) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "code expired"})
				return
			}

			// Mark code as used - one time only
			store.DB.Exec(`UPDATE authorization_codes SET used = 1 WHERE code = ?`, body.Code)

			// Fetch email for jwt claims
			var email string
			store.DB.QueryRow(`SELECT email FROM users WHERE id = ?`, userID).Scan(&email)

			// Issue tokens
			accessToken, _ := signJWT(userID, email, cfg.JWTSecret, 15*time.Minute)
			refreshToken, _ := signJWT(userID, email, cfg.JWTSecret, 6*24*time.Hour)
			store.DB.Exec(
				`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
				refreshToken, userID, time.Now().Add(7*24*time.Hour),
			)

			c.JSON(http.StatusOK, gin.H{
				"access_token":  accessToken,
				"refresh_token": refreshToken,
				"token_type":    "Bearer",
				"expires_in":    900,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
		}
	}
}

// POST /oauth/refresh
func Refresh(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Verify the JWT signature is valid
		token, err := jwt.ParseWithClaims(body.RefreshToken, &Claims{}, func(t *jwt.Token) (any, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// 2. Check it exists in SQLite and isn't revoked
		var revoked int
		var expiresAt time.Time
		err = store.DB.QueryRow(`SELECT revoked, expires_at FROM refresh_tokens WHERE token = ?`, body.RefreshToken).Scan(&revoked, &expiresAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}
		if revoked == 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token revoked"})
			return
		}
		if time.Now().After(expiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		}

		userID := claims.Subject
		email := claims.Email

		// 3. Rotate - revoke old, issue new refresh token
		_, err = store.DB.Exec(
			`UPDATE refresh_tokens SET revoked = 1 WHERE TOKEN = ?`,
			body.RefreshToken,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke token"})
			return
		}

		newRefresh, err := signJWT(userID, email, cfg.JWTSecret, 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue refresh token"})
		}
		store.DB.Exec(
			`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
			newRefresh, userID, time.Now().Add(7*24*time.Hour),
		)

		// 4. Issue new access token
		newAccess, err := signJWT(userID, email, cfg.JWTSecret, 15*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue access token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  newAccess,
			"refresh_token": newRefresh,
			"token_type":    "Bearer",
			"expires_in":    900,
		})
	}
}

// POST /oauth/revoke
func Revoke(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify the JWT is at least structurally valid before revoking
		_, err := jwt.ParseWithClaims(body.RefreshToken, &Claims{}, func(t *jwt.Token) (any, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		result, err := store.DB.Exec(
			`UPDATE refresh_tokens SET revoked = 1 WHERE token = ? AND revoked = 0`,
			body.RefreshToken,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke refresh token"})
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			// Already revoked or never existed -- still return 200
			// Don't leak whether the token existed at all
			c.JSON(http.StatusOK, gin.H{"message": "token revoked"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "token revoked"})
	}
}

// POST /clients
func RegisterClient(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Name        string `json:"name" binding:"required"`
			RedirectURI string `json:"redirect_uri" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		clientID := uuid.NewString()
		clientSecret := uuid.NewString() // Raw - shown once, never again

		hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash secret"})
			return
		}

		_, err = store.DB.Exec(
			`INSERT INTO clients (id, secret, name, redirect_uri) VALUES (?, ?, ?, ?)`,
			clientID, string(hash), body.Name, body.RedirectURI,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create client"})
			return
		}

		// Return the raw secret once -- it cannot be recovered after this
		c.JSON(http.StatusCreated, gin.H{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"name":          body.Name,
			"redirect_uri":  body.RedirectURI,
		})
	}
}

// GET /oauth/authorize?client_id=...&redirect_uri=...&state=...
func Authorize(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.Query("client_id")
		redirectURI := c.Query("redirect_uri")
		state := c.Query("state") // opaque value, echoed back to client

		if clientID == "" || redirectURI == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing client_id or redirect_uri"})
			return
		}

		// Validate client exists and redirect_uri matches
		var storedRedirect string
		err := store.DB.QueryRow(
			`SELECT redirect_uri FROM clients WHERE id = ?`, clientID,
		).Scan(&storedRedirect)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown client"})
			return
		}
		if storedRedirect != redirectURI {
			c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri mismatch"})
			return
		}

		// Serve a minimal login form - in production this would be a real HTML page
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
			<html><body>
			<h2>Login</h2>
			<form method="POST" action="/oauth/authorize">
				<input type="hidden" name="client_id" value="`+clientID+`" />
				<input type="hidden" name="redirect_uri" value="`+redirectURI+`" />
				<input type="hidden" name="state" value="`+state+`" />
				<input type="email" name="email" placeholder="Email" required />
				<input type="password" name="password" placeholder="Password" required />
				<button type="submit">Authorize</button>
			</form>
			</body></html>
		`)
	}
}

// POST /oauth/authorize - form submission
func AuthorizeSubmit(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.PostForm("client_id")
		redirectURI := c.PostForm("redirect_uri")
		state := c.PostForm("state")
		email := c.PostForm("email")
		password := c.PostForm("password")

		// Re-validate client
		var storedRedirect string
		err := store.DB.QueryRow(
			`SELECT redirect_uri FROM clients WHERE id = ?`, clientID,
		).Scan(&storedRedirect)
		if err == sql.ErrNoRows || storedRedirect != redirectURI {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client"})
			return
		}

		// Verify user credentials
		var userID, hashed string
		err = store.DB.QueryRow(
			`SELECT id, password FROM users WHERE email = ?`, email,
		).Scan(&userID, &hashed)
		if err == sql.ErrNoRows {
			c.Redirect(http.StatusFound, redirectURI+"?error=access_denied&state="+state)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
			c.Redirect(http.StatusFound, redirectURI+"?error=access_denied&state="+state)
			return
		}

		// Issue a short-lived authorization code (10 min)
		code := uuid.NewString()
		_, err = store.DB.Exec(
			`INSERT INTO authorization_codes (code, user_id, client_id, expires_at) VALUES (?, ?, ?, ?)`,
			code, userID, clientID, time.Now().Add(10*time.Minute),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue code"})
			return
		}

		// Redirect back the client app with the code
		c.Redirect(http.StatusFound, redirectURI+"?code="+code+"&state="+state)
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
