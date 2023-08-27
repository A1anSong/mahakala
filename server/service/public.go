package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping(c *gin.Context) {
	reqIP := c.ClientIP()
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"ip":      reqIP,
	})
}
