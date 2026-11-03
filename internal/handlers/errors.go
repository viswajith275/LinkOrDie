package handlers

import (
	"errors"
	"link_or_die/internal/apperror"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HTTPrespondError(c *gin.Context, err error) bool {
	var ve *apperror.ValidationError
	if errors.As(err, &ve) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"field":   ve.Field,
			"message": ve.Message,
		})
		return true
	}
	if errors.Is(err, apperror.ErrNotFound) {
		c.String(http.StatusNotFound, "Not found")
		return true
	}
	if errors.Is(err, apperror.ErrConflict) {
		c.String(http.StatusConflict, "Conflict")
		return true
	}
	if errors.Is(err, apperror.ErrUnAuthorized) {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return true
	}
	if errors.Is(err, apperror.ErrInvalidURL) {
		c.String(http.StatusUnprocessableEntity, "Invalid URL")
		return true
	}
	return false
}

func internalError(c *gin.Context, err error) {
	log.Printf("internal error: %v", err)
	c.String(http.StatusInternalServerError, "Something went wrong")
}
