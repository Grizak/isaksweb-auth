package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":    c.GetString("user_id"),
		"email": c.GetString("email"),
	})
}
