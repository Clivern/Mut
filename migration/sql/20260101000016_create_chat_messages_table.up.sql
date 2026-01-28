CREATE TABLE IF NOT EXISTS chat_messages (
	id UUID PRIMARY KEY,
	chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
	kind VARCHAR(20) NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id_created_at ON chat_messages (chat_id, created_at);
