CREATE TABLE IF NOT EXISTS message (
    id VARCHAR(60) PRIMARY KEY,

    group_name VARCHAR(255),
    user_id VARCHAR(60) NOT NULL,
    receptor VARCHAR(255) NOT NULL,

    message TEXT NOT NULL,

    sender VARCHAR(255) NOT NULL,
    sender_id VARCHAR(60) NOT NULL,
    status VARCHAR(50) NOT NULL,

    attempt INT DEFAULT 0,
    last_attempt TIMESTAMP WITH TIME ZONE,
    last_providers TEXT,
    response TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP 
);