CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    send_time TIMESTAMP NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_send_time ON notifications(send_time);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);