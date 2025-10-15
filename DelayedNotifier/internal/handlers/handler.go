package handlers

import (
	"context"
	"fmt"
	"time"

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
	fmt.Println("execute creating")
	h.ns.DB.ExecContext(context.Background(), "INSERT INTO NOTIFICATIONS (STATUS,SENDTIME,USERID) VALUES($1,$2,$3)", "Create", time.Now(), 12)
}

func (h *Handler) DeleteNotificationHandler(c *ginext.Context) {

}
