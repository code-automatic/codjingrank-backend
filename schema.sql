-- Coding Judge Platform standalone PostgreSQL schema.
-- PostgreSQL 16+, shared by Auth, Gateway, Run, Submission, Forum and AI.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Clean bootstrap for the new product. Legacy Subject/Session/Lesson and
-- Repository/GitHub tables are intentionally excluded.
DROP VIEW IF EXISTS leaderboard CASCADE;
DROP TABLE IF EXISTS ai_review_findings CASCADE;
DROP TABLE IF EXISTS ai_reviews CASCADE;
DROP TABLE IF EXISTS daily_challenges CASCADE;
DROP TABLE IF EXISTS forum_post_likes CASCADE;
DROP TABLE IF EXISTS forum_comments CASCADE;
DROP TABLE IF EXISTS forum_posts CASCADE;
DROP TABLE IF EXISTS user_problem_progress CASCADE;
DROP TABLE IF EXISTS accepted_solutions CASCADE;
DROP TABLE IF EXISTS submission_test_results CASCADE;
DROP TABLE IF EXISTS submissions CASCADE;
DROP TABLE IF EXISTS problem_test_cases CASCADE;
DROP TABLE IF EXISTS problem_samples CASCADE;
DROP TABLE IF EXISTS code_templates CASCADE;
DROP TABLE IF EXISTS problem_languages CASCADE;
DROP TABLE IF EXISTS problem_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS problems CASCADE;
DROP TABLE IF EXISTS levels CASCADE;
DROP TABLE IF EXISTS packages CASCADE;
DROP TABLE IF EXISTS languages CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP FUNCTION IF EXISTS set_updated_at() CASCADE;
DROP FUNCTION IF EXISTS prevent_published_test_mutation() CASCADE;
DROP FUNCTION IF EXISTS check_submission_cooldown() CASCADE;
DROP FUNCTION IF EXISTS validate_submission_transition() CASCADE;
DROP FUNCTION IF EXISTS validate_published_package() CASCADE;
DROP FUNCTION IF EXISTS validate_published_problem() CASCADE;
DROP FUNCTION IF EXISTS validate_accepted_solution() CASCADE;
DROP FUNCTION IF EXISTS validate_ai_review() CASCADE;
DROP TYPE IF EXISTS ai_finding_severity CASCADE;
DROP TYPE IF EXISTS ai_review_status CASCADE;
DROP TYPE IF EXISTS forum_post_status CASCADE;
DROP TYPE IF EXISTS forum_post_type CASCADE;
DROP TYPE IF EXISTS test_case_visibility CASCADE;
DROP TYPE IF EXISTS test_result_status CASCADE;
DROP TYPE IF EXISTS submission_status CASCADE;
DROP TYPE IF EXISTS execution_mode CASCADE;
DROP TYPE IF EXISTS difficulty_level CASCADE;
DROP TYPE IF EXISTS problem_status CASCADE;
DROP TYPE IF EXISTS package_status CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;

CREATE TYPE user_role AS ENUM ('student', 'teacher', 'admin');
CREATE TYPE package_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE problem_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');
CREATE TYPE execution_mode AS ENUM ('function', 'stdin_stdout');
CREATE TYPE submission_status AS ENUM (
    'Pending', 'Running', 'Accepted', 'Wrong Answer',
    'Time Limit Exceeded', 'Memory Limit Exceeded', 'Compilation Error',
    'Runtime Error', 'Internal Error'
);
CREATE TYPE test_result_status AS ENUM (
    'Accepted', 'Wrong Answer', 'Time Limit Exceeded',
    'Memory Limit Exceeded', 'Compilation Error', 'Runtime Error',
    'Internal Error'
);
CREATE TYPE test_case_visibility AS ENUM ('sample', 'hidden');
CREATE TYPE forum_post_type AS ENUM ('discussion', 'question', 'solution');
CREATE TYPE forum_post_status AS ENUM ('published', 'hidden', 'deleted');
CREATE TYPE ai_review_status AS ENUM ('queued', 'running', 'completed', 'failed');
CREATE TYPE ai_finding_severity AS ENUM ('info', 'warning', 'error');

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    username text NOT NULL,
    full_name text,
    role user_role NOT NULL DEFAULT 'student',
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_users_email_not_blank CHECK (btrim(email) <> ''),
    CONSTRAINT chk_users_username_not_blank CHECK (btrim(username) <> ''),
    CONSTRAINT chk_users_password_hash_not_blank CHECK (btrim(password_hash) <> '')
);
CREATE UNIQUE INDEX uq_users_email_lower ON users (lower(email));
CREATE UNIQUE INDEX uq_users_username_lower ON users (lower(username));
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    rotated_from uuid REFERENCES refresh_tokens (id) ON DELETE SET NULL,
    user_agent text,
    ip_address inet,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_refresh_token_expiry CHECK (expires_at > created_at),
    CONSTRAINT chk_refresh_token_revoked CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id, created_at DESC);
CREATE INDEX idx_refresh_tokens_expiry ON refresh_tokens (expires_at) WHERE revoked_at IS NULL;

CREATE TABLE packages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    description text,
    status package_status NOT NULL DEFAULT 'draft',
    display_order integer NOT NULL DEFAULT 0,
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_packages_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_packages_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_packages_display_order CHECK (display_order >= 0),
    CONSTRAINT chk_packages_published_at CHECK (status <> 'published' OR published_at IS NOT NULL)
);
CREATE INDEX idx_packages_public_order ON packages (display_order, id) WHERE status = 'published';
CREATE TRIGGER trg_packages_updated_at BEFORE UPDATE ON packages
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE levels (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id uuid NOT NULL REFERENCES packages (id) ON DELETE RESTRICT,
    level_no smallint NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    difficulty difficulty_level NOT NULL,
    display_order smallint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_levels_package_no UNIQUE (package_id, level_no),
    CONSTRAINT uq_levels_package_slug UNIQUE (package_id, slug),
    CONSTRAINT uq_levels_package_difficulty UNIQUE (package_id, difficulty),
    CONSTRAINT chk_levels_no CHECK (level_no BETWEEN 1 AND 3),
    CONSTRAINT chk_levels_order CHECK (display_order BETWEEN 1 AND 3),
    CONSTRAINT chk_levels_difficulty CHECK (
        (level_no = 1 AND difficulty = 'easy') OR
        (level_no = 2 AND difficulty = 'medium') OR
        (level_no = 3 AND difficulty = 'hard')
    ),
    CONSTRAINT chk_levels_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_levels_name CHECK (btrim(name) <> '')
);
CREATE INDEX idx_levels_package_order ON levels (package_id, display_order);
CREATE TRIGGER trg_levels_updated_at BEFORE UPDATE ON levels
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE problems (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    level_id uuid NOT NULL REFERENCES levels (id) ON DELETE RESTRICT,
    slug text NOT NULL UNIQUE,
    title text NOT NULL,
    statement_md text NOT NULL,
    status problem_status NOT NULL DEFAULT 'draft',
    execution_mode execution_mode NOT NULL DEFAULT 'stdin_stdout',
    function_name text,
    time_limit_ms integer NOT NULL DEFAULT 1000,
    memory_limit_mb integer NOT NULL DEFAULT 128,
    output_limit_bytes integer NOT NULL DEFAULT 1048576,
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_problems_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_problems_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_problems_statement CHECK (btrim(statement_md) <> ''),
    CONSTRAINT chk_problems_limits CHECK (time_limit_ms > 0 AND memory_limit_mb > 0 AND output_limit_bytes > 0),
    CONSTRAINT chk_problems_function_name CHECK (execution_mode <> 'function' OR btrim(COALESCE(function_name, '')) <> ''),
    CONSTRAINT chk_problems_published_at CHECK (status <> 'published' OR published_at IS NOT NULL)
);
CREATE INDEX idx_problems_level_status ON problems (level_id, status, created_at);
CREATE TRIGGER trg_problems_updated_at BEFORE UPDATE ON problems
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION validate_published_package()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.status = 'published' AND (
        SELECT COUNT(*) FROM levels WHERE package_id = NEW.id
    ) <> 3 THEN
        RAISE EXCEPTION 'A published package must contain exactly three levels';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_published_package_guard
BEFORE UPDATE OF status ON packages
FOR EACH ROW EXECUTE FUNCTION validate_published_package();

CREATE OR REPLACE FUNCTION validate_published_problem()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.status = 'published' AND (
        NOT EXISTS (SELECT 1 FROM problem_languages WHERE problem_id = NEW.id)
        OR NOT EXISTS (SELECT 1 FROM problem_samples WHERE problem_id = NEW.id)
        OR NOT EXISTS (SELECT 1 FROM problem_test_cases WHERE problem_id = NEW.id AND is_active)
    ) THEN
        RAISE EXCEPTION 'A published problem needs a language, sample, and active hidden test';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_published_problem_guard
BEFORE UPDATE OF status ON problems
FOR EACH ROW EXECUTE FUNCTION validate_published_problem();

CREATE TABLE tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_tags_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_tags_name CHECK (btrim(name) <> '')
);
CREATE TABLE problem_tags (
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES tags (id) ON DELETE RESTRICT,
    PRIMARY KEY (problem_id, tag_id)
);
CREATE INDEX idx_problem_tags_tag ON problem_tags (tag_id, problem_id);

CREATE TABLE languages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    piston_alias text NOT NULL,
    piston_version text NOT NULL,
    file_extension text NOT NULL,
    default_file_name text NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_languages_runtime UNIQUE (piston_alias, piston_version),
    CONSTRAINT chk_languages_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_languages_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_languages_runtime CHECK (btrim(piston_alias) <> '' AND btrim(piston_version) <> '')
);
CREATE TRIGGER trg_languages_updated_at BEFORE UPDATE ON languages
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE problem_languages (
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    language_id uuid NOT NULL REFERENCES languages (id) ON DELETE RESTRICT,
    PRIMARY KEY (problem_id, language_id)
);
CREATE TABLE code_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id uuid NOT NULL,
    language_id uuid NOT NULL,
    starter_code text NOT NULL DEFAULT '',
    wrapper_code text,
    file_name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_template_problem_language FOREIGN KEY (problem_id, language_id)
        REFERENCES problem_languages (problem_id, language_id) ON DELETE CASCADE,
    CONSTRAINT uq_code_template_problem_language UNIQUE (problem_id, language_id),
    CONSTRAINT chk_code_template_file_name CHECK (btrim(file_name) <> '')
);
CREATE TRIGGER trg_code_templates_updated_at BEFORE UPDATE ON code_templates
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE problem_samples (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    input_data text,
    expected_output text NOT NULL,
    display_order integer NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_problem_sample_order UNIQUE (problem_id, display_order),
    CONSTRAINT chk_problem_sample_order CHECK (display_order > 0)
);
CREATE TABLE problem_test_cases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE RESTRICT,
    input_data text,
    expected_output text NOT NULL,
    display_order integer NOT NULL DEFAULT 1,
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_problem_test_case_order UNIQUE (problem_id, display_order),
    CONSTRAINT chk_problem_test_case_order CHECK (display_order > 0)
);
CREATE INDEX idx_problem_test_cases_judge ON problem_test_cases (problem_id, display_order) WHERE is_active;

CREATE OR REPLACE FUNCTION prevent_published_test_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    target_problem uuid;
    target_status problem_status;
BEGIN
    target_problem := COALESCE(NEW.problem_id, OLD.problem_id);
    SELECT status INTO target_status FROM problems WHERE id = target_problem;
    IF target_status = 'published' THEN
        RAISE EXCEPTION 'Test cases are immutable after problem publication';
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$;
CREATE TRIGGER trg_problem_samples_immutable
BEFORE INSERT OR UPDATE OR DELETE ON problem_samples
FOR EACH ROW EXECUTE FUNCTION prevent_published_test_mutation();
CREATE TRIGGER trg_problem_test_cases_immutable
BEFORE INSERT OR UPDATE OR DELETE ON problem_test_cases
FOR EACH ROW EXECUTE FUNCTION prevent_published_test_mutation();

CREATE TABLE submissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE RESTRICT,
    language_id uuid NOT NULL REFERENCES languages (id) ON DELETE RESTRICT,
    source_code text NOT NULL,
    source_hash char(64) NOT NULL,
    status submission_status NOT NULL DEFAULT 'Pending',
    runtime_ms integer,
    memory_kb integer,
    error_message text,
    pass_count integer NOT NULL DEFAULT 0,
    total_testcases integer NOT NULL DEFAULT 0,
    attempt_count integer NOT NULL DEFAULT 0,
    claimed_at timestamptz,
    judged_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_submissions_source_hash CHECK (source_hash ~ '^[0-9a-f]{64}$'),
    CONSTRAINT chk_submissions_metrics CHECK (
        (runtime_ms IS NULL OR runtime_ms >= 0) AND
        (memory_kb IS NULL OR memory_kb >= 0) AND
        pass_count >= 0 AND total_testcases >= 0 AND pass_count <= total_testcases AND
        attempt_count >= 0
    ),
    CONSTRAINT chk_submissions_claim CHECK (status <> 'Running' OR claimed_at IS NOT NULL)
    ,CONSTRAINT fk_submission_problem_language FOREIGN KEY (problem_id, language_id)
        REFERENCES problem_languages (problem_id, language_id)
);
CREATE INDEX idx_submissions_queue ON submissions (status, claimed_at, created_at)
    WHERE status IN ('Pending', 'Running');
CREATE INDEX idx_submissions_user_history ON submissions (user_id, created_at DESC);
CREATE INDEX idx_submissions_problem_history ON submissions (problem_id, created_at DESC);
CREATE TRIGGER trg_submissions_updated_at BEFORE UPDATE ON submissions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION check_submission_cooldown()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM submissions
        WHERE user_id = NEW.user_id
          AND problem_id = NEW.problem_id
          AND created_at > CURRENT_TIMESTAMP - INTERVAL '10 seconds'
    ) THEN
        RAISE EXCEPTION 'Please wait 10 seconds before submitting again';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_submissions_cooldown BEFORE INSERT ON submissions
FOR EACH ROW EXECUTE FUNCTION check_submission_cooldown();

CREATE OR REPLACE FUNCTION validate_submission_transition()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status NOT IN ('Pending', 'Running') AND NEW.status <> OLD.status THEN
        RAISE EXCEPTION 'Final submission verdict cannot change';
    END IF;
    IF OLD.status = 'Pending' AND NEW.status NOT IN ('Pending', 'Running', 'Internal Error') THEN
        RAISE EXCEPTION 'Pending submission must be claimed before finalization';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_submission_status_transition
BEFORE UPDATE OF status ON submissions
FOR EACH ROW EXECUTE FUNCTION validate_submission_transition();

CREATE TABLE submission_test_results (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id uuid NOT NULL REFERENCES submissions (id) ON DELETE CASCADE,
    test_case_id uuid NOT NULL REFERENCES problem_test_cases (id) ON DELETE RESTRICT,
    status test_result_status NOT NULL,
    runtime_ms integer,
    memory_kb integer,
    actual_output text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_submission_test_result UNIQUE (submission_id, test_case_id),
    CONSTRAINT chk_submission_test_metrics CHECK (runtime_ms IS NULL OR runtime_ms >= 0)
);

CREATE TABLE accepted_solutions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id uuid NOT NULL REFERENCES submissions (id) ON DELETE RESTRICT,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE RESTRICT,
    language_id uuid NOT NULL REFERENCES languages (id) ON DELETE RESTRICT,
    source_hash char(64) NOT NULL,
    accepted_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_accepted_solution_source UNIQUE (user_id, problem_id, language_id, source_hash),
    CONSTRAINT uq_accepted_solution_submission UNIQUE (submission_id),
    CONSTRAINT fk_accepted_problem_language FOREIGN KEY (problem_id, language_id)
        REFERENCES problem_languages (problem_id, language_id),
    CONSTRAINT chk_accepted_solution_hash CHECK (source_hash ~ '^[0-9a-f]{64}$')
);
CREATE INDEX idx_accepted_solutions_user ON accepted_solutions (user_id, accepted_at DESC);
CREATE INDEX idx_accepted_solutions_problem ON accepted_solutions (problem_id, accepted_at DESC);

CREATE OR REPLACE FUNCTION validate_accepted_solution()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM submissions
        WHERE id = NEW.submission_id
          AND status = 'Accepted'
          AND user_id = NEW.user_id
          AND problem_id = NEW.problem_id
          AND language_id = NEW.language_id
          AND source_hash = NEW.source_hash
    ) THEN
        RAISE EXCEPTION 'Accepted solution must reference a matching Accepted submission';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_accepted_solution_guard
BEFORE INSERT OR UPDATE ON accepted_solutions
FOR EACH ROW EXECUTE FUNCTION validate_accepted_solution();

CREATE TABLE user_problem_progress (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE RESTRICT,
    first_solved_at timestamptz NOT NULL,
    last_solved_at timestamptz NOT NULL,
    accepted_count integer NOT NULL DEFAULT 1,
    best_runtime_ms integer,
    best_memory_kb integer,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_problem_progress UNIQUE (user_id, problem_id),
    CONSTRAINT chk_progress_count CHECK (accepted_count > 0),
    CONSTRAINT chk_progress_times CHECK (last_solved_at >= first_solved_at),
    CONSTRAINT chk_progress_metrics CHECK (
        (best_runtime_ms IS NULL OR best_runtime_ms >= 0) AND
        (best_memory_kb IS NULL OR best_memory_kb >= 0)
    )
);
CREATE INDEX idx_progress_leaderboard ON user_problem_progress (user_id, accepted_count DESC, first_solved_at);
CREATE INDEX idx_progress_period ON user_problem_progress (first_solved_at, user_id);
CREATE TRIGGER trg_progress_updated_at BEFORE UPDATE ON user_problem_progress
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE VIEW leaderboard AS
SELECT
    u.id AS user_id,
    u.username,
    COUNT(upp.problem_id)::integer AS solved_problem_count,
    COALESCE(SUM(upp.accepted_count), 0)::integer AS accepted_solution_count,
    RANK() OVER (
        ORDER BY COUNT(upp.problem_id) DESC, COALESCE(SUM(upp.accepted_count), 0) DESC
    ) AS rank
FROM users u
JOIN user_problem_progress upp ON upp.user_id = u.id
WHERE u.is_active
GROUP BY u.id, u.username;

CREATE TABLE forum_posts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    problem_id uuid REFERENCES problems (id) ON DELETE SET NULL,
    language_id uuid REFERENCES languages (id) ON DELETE SET NULL,
    post_type forum_post_type NOT NULL,
    status forum_post_status NOT NULL DEFAULT 'published',
    title text NOT NULL,
    content_md text NOT NULL,
    code_snippet text,
    view_count bigint NOT NULL DEFAULT 0,
    like_count bigint NOT NULL DEFAULT 0,
    comment_count bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamptz,
    CONSTRAINT chk_forum_post_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_forum_post_content CHECK (btrim(content_md) <> ''),
    CONSTRAINT chk_forum_post_counters CHECK (view_count >= 0 AND like_count >= 0 AND comment_count >= 0),
    CONSTRAINT chk_forum_post_deleted CHECK (status <> 'deleted' OR deleted_at IS NOT NULL)
);
CREATE INDEX idx_forum_posts_public ON forum_posts (status, created_at DESC);
CREATE INDEX idx_forum_posts_problem ON forum_posts (problem_id, created_at DESC) WHERE status = 'published';
CREATE TRIGGER trg_forum_posts_updated_at BEFORE UPDATE ON forum_posts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE forum_comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id uuid NOT NULL REFERENCES forum_posts (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    parent_id uuid REFERENCES forum_comments (id) ON DELETE CASCADE,
    content text NOT NULL,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_forum_comment_content CHECK (btrim(content) <> ''),
    CONSTRAINT chk_forum_comment_not_self_parent CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE INDEX idx_forum_comments_post ON forum_comments (post_id, created_at);
CREATE TRIGGER trg_forum_comments_updated_at BEFORE UPDATE ON forum_comments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE forum_post_likes (
    post_id uuid NOT NULL REFERENCES forum_posts (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE ai_reviews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id uuid NOT NULL REFERENCES submissions (id) ON DELETE RESTRICT,
    accepted_solution_id uuid REFERENCES accepted_solutions (id) ON DELETE RESTRICT,
    status ai_review_status NOT NULL DEFAULT 'queued',
    score numeric(5, 2),
    summary text,
    dimensions jsonb NOT NULL DEFAULT '{}'::jsonb,
    model text,
    prompt_version text,
    provider_request_id text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at timestamptz,
    completed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_ai_review_submission UNIQUE (submission_id),
    CONSTRAINT chk_ai_review_score CHECK (score IS NULL OR score BETWEEN 0 AND 100),
    CONSTRAINT chk_ai_review_completed CHECK (status NOT IN ('completed', 'failed') OR completed_at IS NOT NULL)
);
CREATE INDEX idx_ai_reviews_queue ON ai_reviews (status, created_at) WHERE status IN ('queued', 'running');
CREATE TRIGGER trg_ai_reviews_updated_at BEFORE UPDATE ON ai_reviews
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION validate_ai_review()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM submissions WHERE id = NEW.submission_id AND status = 'Accepted'
    ) THEN
        RAISE EXCEPTION 'AI review requires an Accepted submission';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_ai_review_guard
BEFORE INSERT OR UPDATE ON ai_reviews
FOR EACH ROW EXECUTE FUNCTION validate_ai_review();

CREATE TABLE ai_review_findings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id uuid NOT NULL REFERENCES ai_reviews (id) ON DELETE CASCADE,
    severity ai_finding_severity NOT NULL,
    line_start integer,
    line_end integer,
    title text NOT NULL,
    explanation text NOT NULL,
    suggestion text,
    display_order integer NOT NULL DEFAULT 1,
    CONSTRAINT uq_ai_finding_order UNIQUE (review_id, display_order),
    CONSTRAINT chk_ai_finding_lines CHECK (
        line_start IS NULL OR (line_start > 0 AND (line_end IS NULL OR line_end >= line_start))
    ),
    CONSTRAINT chk_ai_finding_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_ai_finding_explanation CHECK (btrim(explanation) <> '')
);

CREATE TABLE daily_challenges (
    challenge_date date PRIMARY KEY,
    problem_id uuid NOT NULL REFERENCES problems (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO languages (slug, name, piston_alias, piston_version, file_extension, default_file_name)
VALUES
    ('python', 'Python 3', 'python', '3.10.0', '.py', 'solution.py'),
    ('javascript', 'JavaScript', 'javascript', '18.15.0', '.js', 'solution.js'),
    ('cpp', 'C++', 'cpp', '10.2.0', '.cpp', 'solution.cpp')
ON CONFLICT (slug) DO NOTHING;
