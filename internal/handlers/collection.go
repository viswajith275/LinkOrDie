package handlers

import (
	"link_or_die/internal/domains/collections"
	"link_or_die/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


type CollectionHandler struct {
	svc    *services.CollectionService

}

func CreateCollectionHandler(svc *services.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		svc:    svc,
	}
}

func (h *CollectionHandler) Create (c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	uid := userID.(uuid.UUID)

	var req collections.CreateCollectionRequest

	if err := c.ShouldBind(&req); err != nil {
		// Collection creating with error
		return
	}

	_, err := h.svc.CreateCollection(c, uid, req)

	if err != nil {
		if !HTTPrespondError(c, err) {
			internalError(c, err)
		}
		return
	}

	// return the collection templ component using collection

}

func (h *CollectionHandler) Collection (c *gin.Context) {
	// render collection page
}
