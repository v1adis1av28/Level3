package handlers

import (
	"net/http"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	ns *service.NotificationService
}

func NewHandler(ns *service.NotificationService) *Handler {
	return &Handler{
		ns: ns,
	}
}

func (h *Handler) GetNotificationHandler(c *ginext.Context) {

}

func (h *Handler) CreateNotificationHandler(c *ginext.Context) {
	var notification models.Notification
	err := c.ShouldBindJSON(&notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "error on binding json"})
		return
	}
	if notification.UserId < 0 || len(notification.Text) == 0 {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid json"})
		return
	}

	err = h.ns.CreateNotification(&notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "error on creating notification"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"result": "notification was created", "notification": notification})
}

func (h *Handler) DeleteNotificationHandler(c *ginext.Context) {

}
