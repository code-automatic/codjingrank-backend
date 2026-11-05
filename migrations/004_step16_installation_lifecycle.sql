-- Migration: Add installation lifecycle timestamps
ALTER TABLE github_installations ADD COLUMN IF NOT EXISTS installed_at timestamp;
ALTER TABLE github_installations ADD COLUMN IF NOT EXISTS uninstalled_at timestamp;
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS installed_at timestamp;
ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS uninstalled_at timestamp;
