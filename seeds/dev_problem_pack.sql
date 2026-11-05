-- RinnoGen LeetCode-Style Problem Pack
-- Apply after schema.sql and dev_seed.sql
-- Idempotent: safe to run multiple times

-- ============================================================
-- SESSIONS
-- ============================================================
INSERT INTO subject_sessions (subject_id, title, description, order_index)
VALUES
    (1, 'Session 02 - Arrays and Strings', 'Array manipulation and string algorithms', 2),
    (1, 'Session 03 - Hash Maps', 'Hash map based problems', 3),
    (1, 'Session 04 - Stack', 'Stack data structure problems', 4),
    (1, 'Session 05 - Dynamic Programming Basics', 'Introductory dynamic programming', 5)
ON CONFLICT (subject_id, order_index) DO NOTHING;

-- ============================================================
-- LESSONS
-- ============================================================
INSERT INTO lessons (subject_session_id, title, content_md, order_index)
VALUES
    ((SELECT id FROM subject_sessions WHERE subject_id=1 AND order_index=2), 'Lesson 02 - Array Basics', 'Array traversal, searching, and manipulation', 1),
    ((SELECT id FROM subject_sessions WHERE subject_id=1 AND order_index=2), 'Lesson 03 - String Basics', 'String manipulation and matching', 2),
    ((SELECT id FROM subject_sessions WHERE subject_id=1 AND order_index=3), 'Lesson 04 - Hash Map Problems', 'Using hash maps for efficient lookups', 1),
    ((SELECT id FROM subject_sessions WHERE subject_id=1 AND order_index=4), 'Lesson 05 - Stack Problems', 'Stack-based algorithms', 1),
    ((SELECT id FROM subject_sessions WHERE subject_id=1 AND order_index=5), 'Lesson 06 - DP Basics', 'Introduction to dynamic programming', 1)
ON CONFLICT (subject_session_id, order_index) DO NOTHING;

-- ============================================================
-- TAGS
-- ============================================================
INSERT INTO tags (name) VALUES ('Array'), ('String'), ('Hash Map'), ('Stack'), ('Dynamic Programming'), ('Math'), ('Simulation'), ('Two Pointers'), ('Sliding Window'), ('Divide and Conquer')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- PROBLEM 1: sum-two-numbers (already exists in dev_seed.sql, skip)
-- ============================================================

-- ============================================================
-- PROBLEM 2: contains-duplicate
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'contains-duplicate', 'Contains Duplicate', 'Easy',
    E'## Contains Duplicate\n\nGiven an integer array `nums`, return `true` if any value appears at least twice in the array, and return `false` if every element is distinct.\n\n### Input\nFirst line: integer `n` (size of array)\nSecond line: `n` space-separated integers\n\n### Output\n`true` or `false`\n\n### Examples\n\n**Example 1:**\nInput:\n4\n1 2 3 1\nOutput:\ntrue\n\n**Example 2:**\nInput:\n4\n1 2 3 4\nOutput:\nfalse\n\n### Constraints\n- 1 <= n <= 100000\n- -10^9 <= nums[i] <= 10^9',
    1000, 128, true, 'function', 'containsDuplicate', E'function containsDuplicate(nums) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = containsDuplicate(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'contains-duplicate';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": [[1, 2, 3, 1]]}', 'true', 1),
            (pid, 'sample', '{"args": [[1, 2, 3, 4]]}', 'false', 2),
            (pid, 'hidden', '{"args": [[5]]}', 'false', 3),
            (pid, 'hidden', '{"args": [[1, 1, 1, 2, 2, 3]]}', 'true', 4),
            (pid, 'hidden', '{"args": [[1000000000, -1000000000, 0]]}', 'false', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 1 FROM lessons l, problems p
WHERE l.title = 'Lesson 02 - Array Basics' AND p.slug = 'contains-duplicate'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'contains-duplicate' AND t.name IN ('Array', 'Hash Map')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 3: valid-anagram
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'valid-anagram', 'Valid Anagram', 'Easy',
    E'## Valid Anagram\n\nGiven two strings `s` and `t`, return `true` if `t` is an anagram of `s`, and `false` otherwise.\n\nAn anagram is a word formed by rearranging the letters of another.\n\n### Input\nFirst line: string `s`\nSecond line: string `t`\n\n### Output\n`true` or `false`\n\n### Examples\n\n**Example 1:**\nInput:\nanagram\nnagaram\nOutput:\ntrue\n\n**Example 2:**\nInput:\nrat\ncar\nOutput:\nfalse\n\n### Constraints\n- 1 <= s.length, t.length <= 50000\n- s and t consist of lowercase English letters',
    1000, 128, true, 'function', 'isAnagram', E'function isAnagram(s, t) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = isAnagram(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'valid-anagram';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": ["anagram", "nagaram"]}', 'true', 1),
            (pid, 'sample', '{"args": ["rat", "car"]}', 'false', 2),
            (pid, 'hidden', '{"args": ["a", "b"]}', 'false', 3),
            (pid, 'hidden', '{"args": ["aacc", "ccac"]}', 'false', 4),
            (pid, 'hidden', '{"args": ["listen", "silent"]}', 'true', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 1 FROM lessons l, problems p
WHERE l.title = 'Lesson 03 - String Basics' AND p.slug = 'valid-anagram'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'valid-anagram' AND t.name IN ('String', 'Hash Map')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 4: two-sum
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'two-sum', 'Two Sum', 'Easy',
    E'## Two Sum\n\nGiven an array of integers `nums` and an integer `target`, return indices of the two numbers such that they add up to `target`.\n\nYou may assume that each input has exactly one solution.\n\n### Input\nFirst line: `n target` (n=size, target=sum)\nSecond line: `n` space-separated integers\n\n### Output\nTwo space-separated indices (0-based). Any order is accepted.\n\n### Examples\n\n**Example 1:**\nInput:\n4 9\n2 7 11 15\nOutput:\n0 1\n\n**Example 2:**\nInput:\n3 6\n3 2 4\nOutput:\n1 2\n\n### Constraints\n- 2 <= n <= 10000\n- -10^9 <= nums[i] <= 10^9\n- -10^9 <= target <= 10^9\n- Exactly one valid answer exists',
    1000, 128, true, 'function', 'twoSum', E'function twoSum(nums, target) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = twoSum(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'two-sum';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": [[2, 7, 11, 15], 9]}', '0 1', 1),
            (pid, 'sample', '{"args": [[3, 2, 4], 6]}', '1 2', 2),
            (pid, 'hidden', '{"args": [[5, 5], 10]}', '0 1', 3),
            (pid, 'hidden', '{"args": [[-1, 0, 1, 2, -1], 0]}', '0 2', 4),
            (pid, 'hidden', '{"args": [[1, 2, 3, 8, 9, 11], 14]}', '2 5', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 2 FROM lessons l, problems p
WHERE l.title = 'Lesson 02 - Array Basics' AND p.slug = 'two-sum'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'two-sum' AND t.name IN ('Array', 'Hash Map')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 5: reverse-string
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'reverse-string', 'Reverse String', 'Easy',
    E'## Reverse String\n\nWrite a function that reverses a string. The input string is given as an array of characters.\n\n### Input\nA single line containing a string `s`.\n\n### Output\nThe reversed string.\n\n### Examples\n\n**Example 1:**\nInput:\nhello\nOutput:\nolleh\n\n**Example 2:**\nInput:\nJavaScript\nOutput:\ntpircSavaJ\n\n### Constraints\n- 1 <= s.length <= 100000\n- s consists of printable ASCII characters',
    1000, 128, true, 'function', 'reverseString', E'function reverseString(s) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = reverseString(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'reverse-string';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": ["hello"]}', 'olleh', 1),
            (pid, 'sample', '{"args": ["JavaScript"]}', 'tpircSavaJ', 2),
            (pid, 'hidden', '{"args": ["a"]}', '{"args": ["a"]}', 3),
            (pid, 'hidden', '{"args": ["racecar"]}', '{"args": ["racecar"]}', 4),
            (pid, 'hidden', '{"args": ["12345"]}', '54321', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 2 FROM lessons l, problems p
WHERE l.title = 'Lesson 03 - String Basics' AND p.slug = 'reverse-string'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'reverse-string' AND t.name IN ('String', 'Two Pointers')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 6: palindrome-number
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'palindrome-number', 'Palindrome Number', 'Easy',
    E'## Palindrome Number\n\nGiven an integer `x`, return `true` if `x` is a palindrome, and `false` otherwise.\n\nAn integer is a palindrome when it reads the same backward as forward.\n\n### Input\nA single integer `x`.\n\n### Output\n`true` or `false`\n\n### Examples\n\n**Example 1:**\nInput:\n121\nOutput:\ntrue\n\n**Example 2:**\nInput:\n-121\nOutput:\nfalse\nExplanation: From left to right, it reads -121. From right to left, it becomes 121-. Therefore it is not a palindrome.\n\n**Example 3:**\nInput:\n10\nOutput:\nfalse\n\n### Constraints\n- -2^31 <= x <= 2^31 - 1',
    1000, 128, true, 'function', 'isPalindrome', E'function isPalindrome(x) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = isPalindrome(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'palindrome-number';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": [121]}', 'true', 1),
            (pid, 'sample', '{"args": [-121]}', 'false', 2),
            (pid, 'hidden', '{"args": [10]}', 'false', 3),
            (pid, 'hidden', '{"args": [0]}', 'true', 4),
            (pid, 'hidden', '{"args": [123454321]}', 'true', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 3 FROM lessons l, problems p
WHERE l.title = 'Lesson 03 - String Basics' AND p.slug = 'palindrome-number'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'palindrome-number' AND t.name IN ('Math')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 7: maximum-subarray
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'maximum-subarray', 'Maximum Subarray', 'Medium',
    E'## Maximum Subarray\n\nGiven an integer array `nums`, find the subarray with the largest sum, and return its sum.\n\n### Input\nFirst line: integer `n` (size of array)\nSecond line: `n` space-separated integers\n\n### Output\nThe maximum subarray sum.\n\n### Examples\n\n**Example 1:**\nInput:\n9\n-2 1 -3 4 -1 2 1 -5 4\nOutput:\n6\nExplanation: The subarray [4,-1,2,1] has the largest sum 6.\n\n**Example 2:**\nInput:\n1\n5\nOutput:\n5\n\n### Constraints\n- 1 <= n <= 100000\n- -10^4 <= nums[i] <= 10^4',
    1000, 128, true, 'function', 'maxSubArray', E'function maxSubArray(nums) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = maxSubArray(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'maximum-subarray';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": [[-2, 1, -3, 4, -1, 2, 1, -5, 4]]}', '6', 1),
            (pid, 'sample', '{"args": [[5]]}', '{"args": [5]}', 2),
            (pid, 'hidden', '{"args": [[-1, -2, -3, -4, -5]]}', '-1', 3),
            (pid, 'hidden', '{"args": [[1, 2, 3, 4, 5, 6]]}', '21', 4),
            (pid, 'hidden', '{"args": [[5, -2, 3, 4]]}', '{"args": [10]}', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 1 FROM lessons l, problems p
WHERE l.title = 'Lesson 06 - DP Basics' AND p.slug = 'maximum-subarray'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'maximum-subarray' AND t.name IN ('Array', 'Dynamic Programming', 'Divide and Conquer')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 8: fizz-buzz
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'fizz-buzz', 'Fizz Buzz', 'Easy',
    E'## Fizz Buzz\n\nGiven an integer `n`, return a string array where:\n- `Fizz` for multiples of 3\n- `Buzz` for multiples of 5\n- `FizzBuzz` for multiples of both 3 and 5\n- the number itself as a string for all other cases\n\n### Input\nA single integer `n`.\n\n### Output\n`n` lines, one for each number from 1 to n.\n\n### Examples\n\n**Example 1:**\nInput:\n5\nOutput:\n1\n2\nFizz\n4\nBuzz\n\n**Example 2:**\nInput:\n15\nOutput:\n1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n\n### Constraints\n- 1 <= n <= 10000',
    1000, 128, true, 'function', 'fizzBuzz', E'function fizzBuzz(n) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = fizzBuzz(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'fizz-buzz';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": [5]}', '1\n2\nFizz\n4\nBuzz', 1),
            (pid, 'sample', '{"args": [3]}', '1\n2\nFizz', 2),
            (pid, 'hidden', '{"args": [1]}', '{"args": [1]}', 3),
            (pid, 'hidden', '{"args": [15]}', '1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz', 4),
            (pid, 'hidden', '{"args": [30]}', '1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n16\n17\nFizz\n19\nBuzz\nFizz\n22\n23\nFizz\nBuzz\n26\nFizz\n28\n29\nFizzBuzz', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 3 FROM lessons l, problems p
WHERE l.title = 'Lesson 02 - Array Basics' AND p.slug = 'fizz-buzz'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'fizz-buzz' AND t.name IN ('Math', 'Simulation')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 9: count-vowels
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'count-vowels', 'Count Vowels', 'Easy',
    E'## Count Vowels\n\nGiven a string `s`, return the number of vowels in the string.\n\nVowels are: a, e, i, o, u (both lowercase and uppercase).\n\n### Input\nA single line containing a string `s`.\n\n### Output\nAn integer representing the number of vowels.\n\n### Examples\n\n**Example 1:**\nInput:\nhello\nOutput:\n2\n\n**Example 2:**\nInput:\nJavaScript\nOutput:\n3\n\n### Constraints\n- 1 <= s.length <= 100000\n- s consists of printable ASCII characters',
    1000, 128, true, 'function', 'countVowels', E'function countVowels(s) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = countVowels(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'count-vowels';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": ["hello"]}', '2', 1),
            (pid, 'sample', '{"args": ["JavaScript"]}', '{"args": [3]}', 2),
            (pid, 'hidden', '{"args": ["aeiou"]}', '{"args": [5]}', 3),
            (pid, 'hidden', '{"args": ["AEIOUaeiou"]}', '{"args": [10]}', 4),
            (pid, 'hidden', '{"args": ["bcdfg"]}', '{"args": [0]}', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 4 FROM lessons l, problems p
WHERE l.title = 'Lesson 03 - String Basics' AND p.slug = 'count-vowels'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'count-vowels' AND t.name IN ('String')
ON CONFLICT DO NOTHING;

-- ============================================================
-- PROBLEM 10: valid-parentheses
-- ============================================================
INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, sample_test_cases)
VALUES (
    'valid-parentheses', 'Valid Parentheses', 'Easy',
    E'## Valid Parentheses\n\nGiven a string `s` containing just the characters `(`, `)`, `{`, `}`, `[`, `]`, determine if the input string is valid.\n\nAn input string is valid if:\n1. Open brackets must be closed by the same type of brackets.\n2. Open brackets must be closed in the correct order.\n3. Every close bracket has a corresponding open bracket of the same type.\n\n### Input\nA single line containing a string `s`.\n\n### Output\n`true` or `false`\n\n### Examples\n\n**Example 1:**\nInput:\n()\nOutput:\ntrue\n\n**Example 2:**\nInput:\n()[]{}\nOutput:\ntrue\n\n**Example 3:**\nInput:\n(]\nOutput:\nfalse\n\n### Constraints\n- 1 <= s.length <= 10000\n- s consists of parentheses only `()[]{}`',
    1000, 128, true, 'function', 'isValid', E'function isValid(s) {\n  // write your code here\n}', E'const fs = require("fs");\nconst input = JSON.parse(fs.readFileSync(0, "utf8"));\nconst result = isValid(...input.args);\nconsole.log(JSON.stringify(result));', '[]'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET 
    execution_mode = EXCLUDED.execution_mode,
    function_name = EXCLUDED.function_name,
    initial_code = EXCLUDED.initial_code,
    driver_code = EXCLUDED.driver_code,
    problem_md = EXCLUDED.problem_md;

DO $$
DECLARE
    pid INTEGER;
BEGIN
    SELECT id INTO pid FROM problems WHERE slug = 'valid-parentheses';
    IF NOT EXISTS (SELECT 1 FROM test_cases WHERE problem_id = pid LIMIT 1) THEN
        INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, order_index) VALUES
            (pid, 'sample', '{"args": ["()"]}', 'true', 1),
            (pid, 'sample', '{"args": ["()[]{}"]}', 'true', 2),
            (pid, 'hidden', '{"args": ["(]"]}', 'false', 3),
            (pid, 'hidden', '{"args": ["([)]"]}', 'false', 4),
            (pid, 'hidden', '{"args": ["{(([]))}"]}', 'true', 5);
    END IF;
END $$;

INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
SELECT l.id, p.id, 1 FROM lessons l, problems p
WHERE l.title = 'Lesson 05 - Stack Problems' AND p.slug = 'valid-parentheses'
ON CONFLICT (lesson_id, problem_id) DO NOTHING;

INSERT INTO problem_tags (problem_id, tag_id)
SELECT p.id, t.id FROM problems p, tags t WHERE p.slug = 'valid-parentheses' AND t.name IN ('String', 'Stack')
ON CONFLICT DO NOTHING;
