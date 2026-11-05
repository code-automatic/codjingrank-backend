-- Migration: Add per-user GitHub owner support
-- Run against existing databases before deploying STEP 8

ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS github_owner varchar(255);
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS github_owner_type varchar(20);

-- Backfill: use github_username as github_owner for existing rows
UPDATE github_accounts SET github_owner = COALESCE(github_username, 'unknown') WHERE github_owner IS NULL;
UPDATE github_accounts SET github_owner_type = 'User' WHERE github_owner_type IS NULL;

ALTER TABLE github_accounts ALTER COLUMN github_owner SET NOT NULL;
ALTER TABLE github_accounts ALTER COLUMN github_owner_type SET NOT NULL;
ALTER TABLE github_accounts ADD CONSTRAINT chk_github_owner_type CHECK (github_owner_type IN ('User', 'Organization'));

ALTER TABLE repositories ADD COLUMN IF NOT EXISTS github_owner varchar(255);
