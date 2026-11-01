package middleware

import (
	"net/http"

	"github.com/39george/scs_gin_adapter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthRequired(adapter *scs_gin_adapter.GinAdapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		val := adapter.Get(c, "userID")
		if val == nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		var uid uuid.UUID
		switch v := val.(type) {
		case string:
			parsed, err := uuid.Parse(v)
			if err != nil {
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			uid = parsed
		default:
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Set("userID", uid)
		c.Next()
	}
}
