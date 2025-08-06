CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(26) PRIMARY KEY,
    user_name VARCHAR(20) NOT NULL,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    password VARCHAR(60) NOT NULL,
    salt VARCHAR(60) NOT NULL,
    role_id VARCHAR(26) NOT NULL,
    mobile VARCHAR(11),
    email VARCHAR(50),
    status VARCHAR(20),
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);