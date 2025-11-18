package worker

import (
	"context"
	"time"

	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/wb-go/wbf/retry"
)

func ExpiredBookingsWorker(ctx context.Context, strategy retry.Strategy, interval time.Duration, deadMin int, storage *storage.Storage) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = retry.DoContext(ctx, strategy, func() error {
				tx, err := storage.DB.Master.BeginTx(ctx, nil)
				if err != nil {
					return err
				}
				defer tx.Rollback()

				_, err = tx.ExecContext(ctx,
					`DELETE FROM book WHERE confirmed = false AND created_at < $1`,
					time.Now().Add(-time.Duration(deadMin)*time.Minute),
				)
				return tx.Commit()
			})
		}
	}
}
