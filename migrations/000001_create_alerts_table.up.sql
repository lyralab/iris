CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(100) PRIMARY KEY,
	name  VARCHAR(40) NOT NULL,
    severity VARCHAR(10) NOT NULL,
    description text,
    starts_at timestamp,
    ends_at timestamp,
    status VARCHAR(10)
);