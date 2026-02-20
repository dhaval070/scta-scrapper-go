package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// sendError sends a standardized error response
func sendError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
}
