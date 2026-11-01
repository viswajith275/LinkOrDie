package handlers

import (
	"link_or_die/internals/components"
	"link_or_die/internals/config"
	"link_or_die/internals/domains/dashboard"
	"link_or_die/internals/services"
	"net/http"

	"github.com/39george/scs_gin_adapter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	user    *services.UserService
	cfg     *config.Config
	adapter *scs_gin_adapter.GinAdapter
}

func CreateDashboardHandler(cfg *config.Config, adapter *scs_gin_adapter.GinAdapter, user *services.UserService) *DashboardHandler {
	return &DashboardHandler{
		user:    user,
		cfg:     cfg,
		adapter: adapter,
	}
}

func (d *DashboardHandler) Dashboard(c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	uid := userID.(uuid.UUID)

	user, err := d.user.GetUserById(c.Request.Context(), uid)
	if err != nil {
		if !HTTPrespondError(c, err) {
			internalError(c, err)
		}
		return
	}

	components.Dashboard(dashboard.DashboardView{
		UserName: user.Username,
		Email:    user.Email,
	}).Render(c.Request.Context(), c.Writer)
}
