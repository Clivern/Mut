ALTER TABLE users
	ADD COLUMN provider VARCHAR(20) NOT NULL DEFAULT 'local',
	ADD COLUMN provider_user_id VARCHAR(255),
	ADD COLUMN is_email_verified BOOLEAN DEFAULT false;
