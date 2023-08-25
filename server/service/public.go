package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping(context *gin.Context) {
	reqIP := context.ClientIP()
	context.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"ip":      reqIP,
	})
}
