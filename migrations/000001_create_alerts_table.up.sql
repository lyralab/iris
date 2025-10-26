CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(100) PRIMARY KEY,
	name  VARCHAR(40) NOT NULL,
    severity VARCHAR(10) NOT NULL,
    description text,
    starts_at timestamp,
    ends_at timestamp,
    status VARCHAR(10),
    method VARCHAR(10),
    receptor TEXT,
    send_notif BOOLEAN DEFAULT FALSE,
    silenced INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- trigger function to update updated_at on row update
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- trigger that calls the function before each UPDATE
CREATE TRIGGER set_updated_at_trigger
    BEFORE UPDATE ON alerts
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- optional: index for send_notif = false
CREATE INDEX IF NOT EXISTS idx_alerts_send_notif_false
    ON alerts (id)
    WHERE send_notif = false;
