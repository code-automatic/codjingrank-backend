-- Trusted fixture for a brand new disposable smoke database only.
INSERT INTO packages(id,slug,name,status,published_at)
VALUES('11111111-1111-4111-8111-111111111111','smoke-package','Smoke Package','draft',NULL);
INSERT INTO levels(id,package_id,level_no,difficulty,name,slug,display_order) VALUES
('22222222-2222-4222-8222-222222222221','11111111-1111-4111-8111-111111111111',1,'easy','Easy','easy',1),
('22222222-2222-4222-8222-222222222222','11111111-1111-4111-8111-111111111111',2,'medium','Medium','medium',2),
('22222222-2222-4222-8222-222222222223','11111111-1111-4111-8111-111111111111',3,'hard','Hard','hard',3);
INSERT INTO problems(id,level_id,slug,title,statement_md,time_limit_ms,memory_limit_mb)
VALUES('33333333-3333-4333-8333-333333333333','22222222-2222-4222-8222-222222222221','smoke-sum','Sum','Sum all input integers',1000,128);
INSERT INTO problem_languages(problem_id,language_id)
SELECT '33333333-3333-4333-8333-333333333333',id FROM languages WHERE slug='python';
INSERT INTO code_templates(problem_id,language_id,starter_code,file_name)
SELECT '33333333-3333-4333-8333-333333333333',id,'# starter','main.py' FROM languages WHERE slug='python';
INSERT INTO problem_samples(problem_id,input_data,expected_output)
VALUES('33333333-3333-4333-8333-333333333333','1 2','3');
INSERT INTO problem_test_cases(problem_id,input_data,expected_output,display_order)
VALUES('33333333-3333-4333-8333-333333333333','10 20','30',1),
('33333333-3333-4333-8333-333333333333','-3 8','5',2);
UPDATE problems SET status='published',published_at=now() WHERE slug='smoke-sum';

UPDATE packages SET status='published',published_at=now() WHERE slug='smoke-package';
