package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfFormField = "_csrf"

type csrfStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

var store = &csrfStore{
	tokens: make(map[string]time.Time),
}

func Init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.mu.Lock()
			for token, expiry := range store.tokens {
				if time.Now().After(expiry) {
					delete(store.tokens, token)
				}
			}
			store.mu.Unlock()
		}
	}()
}

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	// Store server-side with 5 min expiry
	store.mu.Lock()
	store.tokens[token] = time.Now().Add(5 * time.Minute)
	store.mu.Unlock()

	return token, nil
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

		store.mu.Lock()
		expiry, exists := store.tokens[formToken]
		if exists {
			delete(store.tokens, formToken) // one-time use
		}
		store.mu.Unlock()

		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unknown CSRF token"})
			return
		}
		if !exists || time.Now().After(expiry) {
			// Preserve the original query params so they don't lose their place
			c.Redirect(http.StatusFound, "/oauth/authorize?client_id="+
				c.PostForm("client_id")+"&redirect_uri="+
				c.PostForm("redirect_uri")+"&state="+
				c.PostForm("state")+"&email="+
				c.PostForm("email"), // Echo email back
			)
			return
		}

		c.Next()
	}
}
