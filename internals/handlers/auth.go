package handlers

import (
	"link_or_die/internals/components"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	components.Landing().Render(c.Request.Context(), c.Writer)
}
