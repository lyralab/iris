CREATE INDEX IF NOT EXISTS idx_message_sender_created_at ON message (sender, created_at DESC);
