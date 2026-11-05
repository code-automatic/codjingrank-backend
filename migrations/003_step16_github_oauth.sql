ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS github_noreply_email varchar(255);
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS commit_author_name varchar(255);
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS oauth_connected_at timestamp;
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS oauth_status varchar(20) NOT NULL DEFAULT 'disconnected';
