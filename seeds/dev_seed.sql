-- RinnoGen MVP Development Seed Data
-- Apply after schema.sql
-- Admin user (password: "password", bcrypt hashed)
INSERT INTO users (email, password, username, full_name, role, is_active)
    VALUES ('admin@example.com', '$2a$10$rHSAVC0IFQOmprSeGM4cDO26pPVZMH/ObzUp0gp2DLlOiYqGnQI6a', 'admin', 'Admin User', 'admin', TRUE)
ON CONFLICT (email)
    DO NOTHING;

-- Published subject
INSERT INTO subjects (title, slug, color, is_published, language_id)
    VALUES ('JavaScript', 'javascript', '#f7df1e', TRUE, 1)
ON CONFLICT (slug)
    DO NOTHING;

-- Session 01
INSERT INTO subject_sessions (subject_id, title, description, order_index)
    VALUES (1, 'Session 01 - Basics', 'Learn the basics of JavaScript programming.', 1)
ON CONFLICT (subject_id, order_index)
    DO NOTHING;

-- Lesson 01
INSERT INTO lessons (subject_session_id, title, content_md, order_index)
    VALUES (1, 'Lesson 01 - Input and Output', 'Learn how to read input and write output in JavaScript.', 1)
ON CONFLICT (subject_session_id, order_index)
    DO NOTHING;

-- Problem: Sum Two Numbers
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code)
    VALUES (
        'sum-two-numbers', 
        'Sum Two Numbers', 
        'Easy', 
        E'## Sum Two Numbers\n\nGiven two integers, return their sum.\n\n### Function Signature\n`function sumTwoNumbers(a, b)`\n\n### Parameters\n- `a` (integer): The first number.\n- `b` (integer): The second number.\n\n### Return Value\n- The sum `a + b`.\n\n### Examples\n\n**Example 1:**\nInput: a = 1, b = 2\nOutput: 3', 
        1000, 
        128, 
        TRUE,
        'function',
        'sumTwoNumbers',
        E'function sumTwoNumbers(a, b) {\n  // write your code here\n}',
        E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = sumTwoNumbers(...input.args);\nconsole.log(JSON.stringify(result));'
    )
ON CONFLICT (slug)
    DO UPDATE SET 
        execution_mode = EXCLUDED.execution_mode,
        function_name = EXCLUDED.function_name,
        initial_code = EXCLUDED.initial_code,
        driver_code = EXCLUDED.driver_code,
        problem_md = EXCLUDED.problem_md;

-- Link lesson to problem
INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
    VALUES (1, 1, 1)
ON CONFLICT (lesson_id, problem_id)
    DO NOTHING;

-- Sample test case (visible to users)
INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index)
    VALUES (1, 'sample', '{"args": [1, 2]}', '3', 1)
ON CONFLICT DO NOTHING;

-- Hidden test case (not visible to users)
INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index)
    VALUES (1, 'hidden', '{"args": [5, 7]}', '12', 2)
ON CONFLICT DO NOTHING;

-- GitHub account — uncomment and fill in your real installation ID
-- Find it at: https://github.com/settings/apps/rinnogen/installations
-- INSERT INTO github_accounts (user_id, installation_id, github_owner, github_owner_type)
-- VALUES (1, '133822192', 'Test-GEM-Support', 'Organization');
