package handlers

import (
	dashboardPage "link_or_die/internal/components/dashboard"
	"link_or_die/internal/domains/dashboard"
	"link_or_die/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	user    *services.UserService
}

func CreateDashboardHandler(user *services.UserService) *DashboardHandler {
	return &DashboardHandler{
		user:    user,
	}
}

func (d *DashboardHandler) Dashboard(c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	uid := userID.(uuid.UUID)

	user, err := d.user.GetUserById(c, uid)
	if err != nil {
		if !HTTPrespondError(c, err) {
			internalError(c, err)
		}
		return
	}

	dashboardPage.Page(dashboard.DashboardView{
		UserName: user.Username,
		Email:    user.Email,
	}).Render(c.Request.Context(), c.Writer)
}
