package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfFormField = "_csrf"

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetCSRFToken(c *gin.Context) (string, error) {
	token, err := GenerateCSRFToken()
	if err != nil {
		return "", err
	}

	c.SetCookie(
		csrfCookieName,
		token,
		300, // 5 min - matches auth code expiry
		"/",
		"",
		os.Getenv("GIN_MODE") == "release", // HTTPS only
		true,                               // HTTPOnly
	)

	return token, nil
}

func ValidateCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(csrfCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing CSRF cookie"})
			return
		}

		formToken := c.PostForm(csrfFormField)
		if formToken == "" || formToken != cookie {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid CSRF token"})
			return
		}

		c.Next()
	}
}
