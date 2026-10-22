ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false;
-- The default admin's generated password was written to the logs.
UPDATE users SET must_change_password = true WHERE user_name = 'admin';
