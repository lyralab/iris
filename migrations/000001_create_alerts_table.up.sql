CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(100) PRIMARY KEY,
	name  VARCHAR(40) NOT NULL,
    severity VARCHAR(10) NOT NULL,
    description text,
    starts_at timestamp,
    ends_at timestamp,
    status VARCHAR(10),
    method VARCHAR(10),
    receptor VARCHAR(100),
    send_notif BOOLEAN DEFAULT FALSE,
    silenced INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_alerts_send_notif_false
    ON alerts (id)
    WHERE send_notif = false;
