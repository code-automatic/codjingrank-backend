\set ON_ERROR_STOP on
BEGIN;
CREATE FUNCTION pg_temp.rejects(statement text) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
    BEGIN
        EXECUTE statement;
    EXCEPTION WHEN OTHERS THEN
        RETURN;
    END;
    RAISE EXCEPTION 'Expected rejection: %', statement;
END;
$$;
INSERT INTO users (id,email,username,password_hash) VALUES ('00000000-0000-0000-0000-000000000001','test@example.org','tester','hash');
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE avatar_url IS NOT NULL OR bio IS NOT NULL) THEN
        RAISE EXCEPTION 'Optional profile fields must default to NULL';
    END IF;
END $$;
UPDATE users SET avatar_url='https://example.org/avatar.png',bio='Practising algorithms' WHERE id='00000000-0000-0000-0000-000000000001';
DO $$ BEGIN
    IF (SELECT avatar_url FROM users WHERE id='00000000-0000-0000-0000-000000000001') IS DISTINCT FROM 'https://example.org/avatar.png'
        OR (SELECT bio FROM users WHERE id='00000000-0000-0000-0000-000000000001') IS DISTINCT FROM 'Practising algorithms' THEN
        RAISE EXCEPTION 'Profile fields did not round-trip';
    END IF;
END $$;
UPDATE users SET avatar_url=repeat('a',2048),bio=repeat('ệ',1000) WHERE id='00000000-0000-0000-0000-000000000001';
UPDATE users SET avatar_url=NULL,bio=NULL WHERE id='00000000-0000-0000-0000-000000000001';
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE avatar_url IS NOT NULL OR bio IS NOT NULL) THEN
        RAISE EXCEPTION 'Optional profile fields must be clearable to NULL';
    END IF;
END $$;
SELECT pg_temp.rejects($q$UPDATE users SET avatar_url=repeat('a',2049)$q$);
SELECT pg_temp.rejects($q$UPDATE users SET bio=repeat('a',1001)$q$);
INSERT INTO packages (id,slug,name) VALUES ('00000000-0000-0000-0000-000000000002','arrays','Arrays');
SELECT pg_temp.rejects($q$UPDATE packages SET status='published',published_at=now()$q$);
INSERT INTO levels (id,package_id,level_no,slug,name,difficulty,display_order)
SELECT ('00000000-0000-0000-0000-00000000000' || (n+2))::uuid,'00000000-0000-0000-0000-000000000002',n,slug,slug,slug::difficulty_level,n FROM (VALUES (1,'easy'),(2,'medium'),(3,'hard')) AS l(n,slug);
UPDATE packages SET status='published',published_at=now();
SELECT pg_temp.rejects($q$INSERT INTO packages (slug,name,status,published_at) VALUES ('broken','Broken','published',now())$q$);
INSERT INTO problems (id,level_id,slug,title,statement_md) VALUES ('00000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000003','sum','Sum','Statement');
SELECT pg_temp.rejects($q$UPDATE problems SET status='published',published_at=now()$q$);
INSERT INTO problem_languages SELECT '00000000-0000-0000-0000-000000000006',id FROM languages WHERE slug='python';
INSERT INTO problem_samples (problem_id,expected_output) VALUES ('00000000-0000-0000-0000-000000000006','1');
INSERT INTO problem_test_cases (problem_id,expected_output) VALUES ('00000000-0000-0000-0000-000000000006','1');
UPDATE problems SET status='published',published_at=now();
SELECT pg_temp.rejects($q$UPDATE problem_test_cases SET expected_output='2'$q$);
INSERT INTO submissions (id,user_id,problem_id,language_id,source_code,source_hash)
SELECT '00000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000006',id,'print(1)',encode(digest('print(1)','sha256'),'hex') FROM languages WHERE slug='python';
SELECT pg_temp.rejects($q$INSERT INTO ai_reviews (submission_id) VALUES ('00000000-0000-0000-0000-000000000007')$q$);
UPDATE submissions SET status='Running',claimed_at=now(),attempt_count=1;
UPDATE submissions SET status='Accepted',pass_count=1,total_testcases=1,judged_at=now();
INSERT INTO accepted_solutions (submission_id,user_id,problem_id,language_id,source_hash) SELECT id,user_id,problem_id,language_id,source_hash FROM submissions;
SELECT pg_temp.rejects($q$INSERT INTO accepted_solutions (submission_id,user_id,problem_id,language_id,source_hash) SELECT id,user_id,problem_id,language_id,source_hash FROM submissions$q$);
INSERT INTO ai_reviews (submission_id) SELECT id FROM submissions;
UPDATE ai_reviews SET status='failed',completed_at=now();
DO $$ BEGIN IF (SELECT status FROM submissions LIMIT 1) <> 'Accepted' THEN RAISE EXCEPTION 'AI failure changed verdict'; END IF; END $$;
SELECT pg_temp.rejects($q$UPDATE submissions SET status='Wrong Answer'$q$);
UPDATE problems SET status='archived';
SELECT pg_temp.rejects($q$UPDATE problem_test_cases SET expected_output='2'$q$);
ROLLBACK;
\echo 'Schema constraint checks passed'
