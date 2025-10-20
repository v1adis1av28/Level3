package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/service"
)

// mock DB
type mockDB struct {
	queryRow func(ctx context.Context, q string, args ...any) *sql.Row
	exec     func(ctx context.Context, q string, args ...any) (sql.Result, error)
}

func (m *mockDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	if m.queryRow != nil {
		return m.queryRow(ctx, q, args...)
	}
	return nil
}

func (m *mockDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	if m.exec != nil {
		return m.exec(ctx, q, args...)
	}
	return nil, nil
}

type mockPublisher struct {
	PublishFunc func(msg []byte, key, ctype string, opt interface{}) error
}

func (m *mockPublisher) Publish(msg []byte, key, ctype string, opt interface{}) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(msg, key, ctype, opt)
	}
	return nil
}

func TestGetRetryDelay(t *testing.T) {
	ns := &service.NotificationService{}
	delay1 := ns.GetRetryDelay(1)
	delay2 := ns.GetRetryDelay(3)
	assert.Equal(t, 30*time.Second, delay1)
	assert.True(t, delay2 > delay1)
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	ns := &service.NotificationService{}
	err := ns.ProcessMessage([]byte("not json"))
	assert.Error(t, err)
}

func TestSendNotification_SometimesFails(t *testing.T) {
	ns := &service.NotificationService{}
	msg := &models.NotificationMessage{NotificationID: 1, UserID: 2, Attempt: 1}
	err := ns.SendNotification(msg)
	if err != nil && !errors.Is(err, nil) {
		t.Logf("send failed as expected: %v", err)
	}
}
