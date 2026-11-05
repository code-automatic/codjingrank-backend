-- Migration: GitHub webhook support tables and columns (STEP 12)

CREATE TABLE IF NOT EXISTS github_installations (
    id serial PRIMARY KEY,
    installation_id varchar(255) UNIQUE NOT NULL,
    github_owner varchar(255) NOT NULL,
    github_owner_type varchar(20) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_github_installations_owner_type CHECK (github_owner_type IN ('User', 'Organization'))
);

DO $$ BEGIN
    CREATE TRIGGER trg_github_installations_updated_at
        BEFORE UPDATE ON github_installations
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
EXCEPTION WHEN duplicate_object THEN null;
END $$;

ALTER TABLE github_accounts ADD COLUMN IF NOT EXISTS status varchar(20) NOT NULL DEFAULT 'active';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS status varchar(20) NOT NULL DEFAULT 'active';
