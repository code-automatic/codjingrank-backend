-- =========================================================================
-- RinnoGen Schema
-- PostgreSQL
-- =========================================================================
-- =========================================================================
-- 1. EXTENSIONS
-- =========================================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================================================================
-- 2. DROP OLD OBJECTS
-- =========================================================================
DROP TABLE IF EXISTS submission_commits CASCADE;

DROP TABLE IF EXISTS submissions CASCADE;

DROP TABLE IF EXISTS lesson_problems CASCADE;

DROP TABLE IF EXISTS problem_tags CASCADE;

DROP TABLE IF EXISTS test_cases CASCADE;

DROP TABLE IF EXISTS problem_language_templates CASCADE;

DROP TABLE IF EXISTS lessons CASCADE;

DROP TABLE IF EXISTS subject_sessions CASCADE;

DROP TABLE IF EXISTS repositories CASCADE;

DROP TABLE IF EXISTS subjects CASCADE;

DROP TABLE IF EXISTS problems CASCADE;

DROP TABLE IF EXISTS languages CASCADE;

DROP TABLE IF EXISTS tags CASCADE;

DROP TABLE IF EXISTS github_accounts CASCADE;

DROP TABLE IF EXISTS users CASCADE;

DROP TYPE IF EXISTS user_role CASCADE;

DROP TYPE IF EXISTS problem_difficulty CASCADE;

DROP TYPE IF EXISTS submission_status CASCADE;

DROP TYPE IF EXISTS test_case_visibility CASCADE;

-- =========================================================================
-- 3. ENUM TYPES
-- =========================================================================
CREATE TYPE user_role AS ENUM (
    'student',
    'teacher',
    'admin'
);

CREATE TYPE problem_difficulty AS ENUM (
    'Easy',
    'Medium',
    'Hard'
);

CREATE TYPE submission_status AS ENUM (
    'Pending',
    'Running',
    'Accepted',
    'Wrong Answer',
    'Time Limit Exceeded',
    'Memory Limit Exceeded',
    'Compilation Error',
    'Runtime Error',
    'Internal Error'
);

CREATE TYPE test_case_visibility AS ENUM (
    'sample',
    'hidden'
);

-- =========================================================================
-- 4. SHARED UPDATED_AT FUNCTION
-- =========================================================================
CREATE OR REPLACE FUNCTION set_updated_at ()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- =========================================================================
-- 5. LEVEL 0 TABLES
-- =========================================================================
CREATE TABLE users (
    id serial PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    username varchar(255) UNIQUE,
    full_name varchar(255),
    ROLE user_role NOT NULL DEFAULT 'student',
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id int NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    user_agent text,
    ip_address text,
    expires_at timestamp NOT NULL,
    revoked_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rotated_from uuid
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

CREATE TABLE tags (
    id serial PRIMARY KEY,
    name varchar(255) UNIQUE NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE languages (
    id serial PRIMARY KEY,
    name varchar(50) NOT NULL,
    piston_alias varchar(50) NOT NULL,
    piston_version varchar(50) NOT NULL,
    file_extension varchar(20),
    default_file_name varchar(100),
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_languages_piston UNIQUE (piston_alias, piston_version)
);

CREATE TRIGGER trg_languages_updated_at
    BEFORE UPDATE ON languages
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

CREATE TABLE problems (
    id serial PRIMARY KEY,
    slug varchar(255) UNIQUE NOT NULL,
    title varchar(255) NOT NULL,
    difficulty problem_difficulty NOT NULL,
    problem_md text NOT NULL,
    time_limit_ms int NOT NULL DEFAULT 1000,
    memory_limit_mb int NOT NULL DEFAULT 128,
    acceptance_rate DECIMAL(5, 2) NOT NULL DEFAULT 0,
    is_published boolean NOT NULL DEFAULT FALSE,
    execution_mode varchar(50) NOT NULL DEFAULT 'function',
    function_name varchar(255),
    initial_code text,
    driver_code text,
    solution_file_name varchar(255),
    -- Dùng cho nút Run nhanh.
    -- Full testcase vẫn nằm ở bảng test_cases.
    sample_test_cases jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_problem_time_limit CHECK (time_limit_ms > 0),
    CONSTRAINT chk_problem_memory_limit CHECK (memory_limit_mb > 0),
    CONSTRAINT chk_problem_acceptance_rate CHECK (acceptance_rate >= 0 AND acceptance_rate <= 100)
);

CREATE TRIGGER trg_problems_updated_at
    BEFORE UPDATE ON problems
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

-- =========================================================================
-- 6. GITHUB ACCOUNT / REPOSITORY TABLES
-- =========================================================================
CREATE TABLE github_accounts (
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    installation_id varchar(255) NOT NULL,
    github_user_id varchar(255),
    github_username varchar(255),
    github_avatar_url text,
    github_owner varchar(255) NOT NULL,
    github_owner_type varchar(20) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'active',
    github_noreply_email varchar(255),
    commit_author_name varchar(255),
    oauth_connected_at timestamp,
    oauth_status varchar(20) NOT NULL DEFAULT 'disconnected',
    installed_at timestamp,
    uninstalled_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_github_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_github_user UNIQUE (user_id),
    CONSTRAINT uq_github_installation UNIQUE (installation_id),
    CONSTRAINT chk_github_owner_type CHECK (github_owner_type IN ('User', 'Organization'))
);

CREATE TRIGGER trg_github_accounts_updated_at
    BEFORE UPDATE ON github_accounts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

-- =========================================================================
-- 6b. GITHUB INSTALLATIONS TABLE (webhook data)
-- =========================================================================
CREATE TABLE github_installations (
    id serial PRIMARY KEY,
    installation_id varchar(255) UNIQUE NOT NULL,
    github_owner varchar(255) NOT NULL,
    github_owner_type varchar(20) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    installed_at timestamp,
    uninstalled_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_github_installations_owner_type CHECK (github_owner_type IN ('User', 'Organization'))
);

CREATE TRIGGER trg_github_installations_updated_at
    BEFORE UPDATE ON github_installations
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

-- =========================================================================
-- 7. CURRICULUM TABLES
-- =========================================================================
CREATE TABLE subjects (
    id serial PRIMARY KEY,
    title varchar(255) NOT NULL,
    slug varchar(255) UNIQUE NOT NULL,
    color varchar(10),
    is_published boolean NOT NULL DEFAULT FALSE,
    -- Nếu một môn chỉ dùng cố định một ngôn ngữ.
    -- Ví dụ JavaScript subject chỉ dùng JavaScript.
    language_id int,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_subject_language FOREIGN KEY (language_id) REFERENCES languages (id)
);

CREATE TRIGGER trg_subjects_updated_at
    BEFORE UPDATE ON subjects
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

CREATE TABLE repositories (
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    subject_id int NOT NULL,
    repo_name varchar(255) NOT NULL,
    repo_full_name varchar(255),
    repo_url text,
    github_repo_id varchar(255),
    github_owner varchar(255),
    status varchar(20) NOT NULL DEFAULT 'active',
    default_branch varchar(100) NOT NULL DEFAULT 'main',
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_repo_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_repo_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE,
    -- 1 user + 1 subject = 1 repo.
    CONSTRAINT uq_repositories_user_subject UNIQUE (user_id, subject_id)
);

CREATE TRIGGER trg_repositories_updated_at
    BEFORE UPDATE ON repositories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

CREATE TABLE subject_sessions (
    id serial PRIMARY KEY,
    subject_id int NOT NULL,
    title varchar(255) NOT NULL,
    description text,
    order_index int NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_session_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE,
    CONSTRAINT uq_subject_session_order UNIQUE (subject_id, order_index)
);

CREATE TRIGGER trg_subject_sessions_updated_at
    BEFORE UPDATE ON subject_sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

CREATE TABLE lessons (
    id serial PRIMARY KEY,
    subject_session_id int NOT NULL,
    title varchar(255) NOT NULL,
    content_md text,
    order_index int NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_lesson_session FOREIGN KEY (subject_session_id) REFERENCES subject_sessions (id) ON DELETE CASCADE,
    CONSTRAINT uq_lesson_order UNIQUE (subject_session_id, order_index)
);

CREATE TRIGGER trg_lessons_updated_at
    BEFORE UPDATE ON lessons
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

-- =========================================================================
-- 8. PROBLEM LANGUAGE TEMPLATE
-- =========================================================================
-- Dùng để mỗi problem có starter/wrapper code riêng theo từng language.
-- Ví dụ JS wrapper khác Python wrapper.
CREATE TABLE problem_language_templates (
    id serial PRIMARY KEY,
    problem_id int NOT NULL,
    language_id int NOT NULL,
    starter_code text,
    wrapper_code text,
    file_name varchar(255),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_plt_problem FOREIGN KEY (problem_id) REFERENCES problems (id) ON DELETE CASCADE,
    CONSTRAINT fk_plt_language FOREIGN KEY (language_id) REFERENCES languages (id),
    CONSTRAINT uq_problem_language_template UNIQUE (problem_id, language_id)
);

CREATE TRIGGER trg_problem_language_templates_updated_at
    BEFORE UPDATE ON problem_language_templates
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at ();

-- =========================================================================
-- 9. TEST CASES
-- =========================================================================
CREATE TABLE test_cases (
    id serial PRIMARY KEY,
    problem_id int NOT NULL,
    visibility test_case_visibility NOT NULL DEFAULT 'hidden',
    -- input_data có thể NULL nếu bài dạng function call bằng execute_code.
    input_data text,
    expected_output text NOT NULL,
    -- execute_code dùng để bọc user code theo kiểu LeetCode.
    -- Có thể NULL nếu bài dạng stdin/stdout thuần.
    execute_code text,
    order_index int NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_test_case_problem FOREIGN KEY (problem_id) REFERENCES problems (id) ON DELETE CASCADE
);

-- =========================================================================
-- 10. MANY-TO-MANY TABLES
-- =========================================================================
CREATE TABLE problem_tags (
    problem_id int NOT NULL,
    tag_id int NOT NULL,
    PRIMARY KEY (problem_id, tag_id),
    CONSTRAINT fk_problem_tag_problem FOREIGN KEY (problem_id) REFERENCES problems (id) ON DELETE CASCADE,
    CONSTRAINT fk_problem_tag_tag FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);

CREATE TABLE lesson_problems (
    lesson_id int NOT NULL,
    problem_id int NOT NULL,
    order_index int NOT NULL,
    PRIMARY KEY (lesson_id, problem_id),
    CONSTRAINT fk_lp_lesson FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE,
    CONSTRAINT fk_lp_problem FOREIGN KEY (problem_id) REFERENCES problems (id) ON DELETE CASCADE,
    CONSTRAINT uq_lesson_problem_order UNIQUE (lesson_id, order_index)
);

-- =========================================================================
-- 11. SUBMISSIONS
-- =========================================================================
CREATE TABLE submissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id int NOT NULL,
    problem_id int NOT NULL,
    language_id int NOT NULL,
    code text NOT NULL,
    status submission_status NOT NULL DEFAULT 'Pending',
    runtime_ms int,
    memory_kb int,
    error_message text,
    pass_count int NOT NULL DEFAULT 0,
    total_testcases int NOT NULL DEFAULT 0,
    -- Có thể giữ nhanh để query dễ.
    -- Chi tiết commit nằm ở submission_commits.
    repo_path text,
    commit_sha varchar(40),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    judged_at timestamp,
    CONSTRAINT fk_submission_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_submission_problem FOREIGN KEY (problem_id) REFERENCES problems (id),
    CONSTRAINT fk_submission_language FOREIGN KEY (language_id) REFERENCES languages (id),
    CONSTRAINT chk_submission_runtime CHECK (runtime_ms IS NULL OR runtime_ms >= 0),
    CONSTRAINT chk_submission_memory CHECK (memory_kb IS NULL OR memory_kb >= 0),
    CONSTRAINT chk_submission_pass_count CHECK (pass_count >= 0),
    CONSTRAINT chk_submission_total_testcases CHECK (total_testcases >= 0),
    CONSTRAINT chk_submission_pass_total CHECK (pass_count <= total_testcases)
);

CREATE TABLE submission_commits (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    submission_id uuid NOT NULL,
    repository_id int NOT NULL,
    file_path text NOT NULL,
    commit_sha varchar(40) NOT NULL,
    commit_url text,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_submission_commit_submission FOREIGN KEY (submission_id) REFERENCES submissions (id) ON DELETE CASCADE,
    CONSTRAINT fk_submission_commit_repository FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE,
    CONSTRAINT uq_submission_commit UNIQUE (submission_id, repository_id, file_path)
);

-- =========================================================================
-- 12. INDEXES
-- =========================================================================
CREATE INDEX idx_subjects_language_id ON subjects (language_id);

CREATE INDEX idx_subject_sessions_subject_order ON subject_sessions (subject_id, order_index);

CREATE INDEX idx_lessons_session_order ON lessons (subject_session_id, order_index);

CREATE INDEX idx_lesson_problems_lesson_order ON lesson_problems (lesson_id, order_index);

CREATE INDEX idx_lesson_problems_problem_id ON lesson_problems (problem_id);

CREATE INDEX idx_test_cases_problem_visibility ON test_cases (problem_id, visibility, order_index);

CREATE INDEX idx_problem_tags_tag_id ON problem_tags (tag_id);

CREATE INDEX idx_repositories_user_id ON repositories (user_id);

CREATE INDEX idx_repositories_subject_id ON repositories (subject_id);

CREATE INDEX idx_submissions_user_problem_created ON submissions (user_id, problem_id, created_at DESC);

CREATE INDEX idx_submissions_problem_status ON submissions (problem_id, status);

CREATE INDEX idx_submissions_user_created ON submissions (user_id, created_at DESC);

CREATE INDEX idx_submission_commits_submission_id ON submission_commits (submission_id);

CREATE INDEX idx_submission_commits_repository_id ON submission_commits (repository_id);

-- =========================================================================
-- 13. TRIGGER: ANTI-SPAM SUBMISSION
-- =========================================================================
-- MVP thì để DB chặn cũng được.
-- Sau này nên chuyển phần này lên API Gateway / service rate limit.
CREATE OR REPLACE FUNCTION check_submission_spam ()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF EXISTS (
        SELECT
            1
        FROM
            submissions
        WHERE
            user_id = NEW.user_id
            AND problem_id = NEW.problem_id
            AND created_at > (NOW() - INTERVAL '10 seconds')) THEN
    RAISE EXCEPTION 'Please wait 10 seconds before submitting again.';
END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_before_submission_insert
    BEFORE INSERT ON submissions
    FOR EACH ROW
    EXECUTE FUNCTION check_submission_spam ();

-- =========================================================================
-- 14. TRIGGER: UPDATE ACCEPTANCE RATE
-- =========================================================================
-- Chỉ tính các submission đã judge xong, không tính Pending/Running.
CREATE OR REPLACE FUNCTION update_acceptance_rate ()
    RETURNS TRIGGER
    AS $$
BEGIN
    UPDATE
        problems
    SET
        acceptance_rate = (
            SELECT
                COALESCE((COUNT(*) FILTER (WHERE status = 'Accepted')::DECIMAL / NULLIF (COUNT(*) FILTER (WHERE status NOT IN ('Pending', 'Running', 'Internal Error')), 0)) * 100, 0)
            FROM
                submissions
            WHERE
                problem_id = NEW.problem_id)
    WHERE
        id = NEW.problem_id;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_after_submission_insert_update_rate
    AFTER INSERT ON submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_acceptance_rate ();

CREATE TRIGGER trg_after_submission_update_rate
    AFTER UPDATE OF status ON submissions
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE FUNCTION update_acceptance_rate ();

-- =========================================================================
-- 15. OPTIONAL SEED DATA
-- =========================================================================
INSERT INTO languages (name, piston_alias, piston_version, file_extension, default_file_name)
VALUES
    ('JavaScript', 'javascript', '18.15.0', '.js', 'solution.js'),
    ('Python 3', 'python', '3.10.0', '.py', 'solution.py'),
    ('C++', 'cpp', '10.2.0', '.cpp', 'solution.cpp')
ON CONFLICT (piston_alias, piston_version)
    DO NOTHING;

CREATE TABLE daily_challenges (
    id SERIAL PRIMARY KEY,
    problem_id INT NOT NULL REFERENCES problems(id),
    challenge_date DATE UNIQUE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
