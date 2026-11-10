#!/usr/bin/env python3
"""Disposable full-stack check. Requires Docker Compose >=2.24.4, Linux cgroup v2.

Runs only trusted fixtures inside Piston. Creates its own Compose project/volumes
and removes them on exit; never uses an existing database or host runtime.
"""
import http.cookiejar
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[2]
PROJECT = 'codjingrank-smoke-' + uuid.uuid4().hex[:10]
ENV = {**os.environ, 'JWT_SECRET': uuid.uuid4().hex + uuid.uuid4().hex}


def request(base, path, payload=None, token=None, expected=200, opener=None, timeout=40):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    req = urllib.request.Request(base + path, headers=headers,
        data=None if payload is None else json.dumps(payload).encode())
    try:
        response = (opener or urllib.request.build_opener()).open(req, timeout=timeout)
    except urllib.error.HTTPError as err:
        response = err
    with response:
        body = response.read()
        assert response.status == expected, f'{path}: HTTP {response.status}, expected {expected}'
        return json.loads(body) if body else None


def main():
    with tempfile.TemporaryDirectory(prefix=PROJECT) as temp:
        override = Path(temp) / 'compose.yaml'
        override.write_text('''services:
  api_gateway:
    ports: !override ["127.0.0.1::8080"]
  piston:
    restart: "no"
    ports: !override ["127.0.0.1::2000"]
  auth_service:
    ports: !reset []
  run_service:
    ports: !reset []
  submission_service:
    ports: !reset []
  postgres:
    ports: !reset []
  redis:
    ports: !reset []
''')
        compose = ['docker', 'compose', '--env-file', '/dev/null', '-p', PROJECT,
                   '-f', str(ROOT / 'docker-compose.yml'), '-f', str(override)]

        def run(*args, stdin=None):
            return subprocess.run(compose + list(args), cwd=ROOT, env=ENV,
                input=stdin, text=True, stdout=subprocess.PIPE,
                stderr=subprocess.PIPE, timeout=600, check=True).stdout.strip()

        def address(service, port):
            return 'http://' + run('port', service, str(port))

        def sql(statement):
            return run('exec', '-T', 'postgres', 'psql', '-U', 'innogen',
                       '-d', 'innogen', '-At', '-v', 'ON_ERROR_STOP=1', stdin=statement)

        def redis(*args):
            return run('exec', '-T', 'redis', 'redis-cli', '--raw', *args)

        assert not run('ps', '-aq'), 'Refusing to reuse a Compose project'
        try:
            print('Building and starting isolated stack', flush=True)
            run('up', '-d', '--build', '--wait', '--wait-timeout', '180')
            gateway, piston = address('api_gateway', 8080), address('piston', 2000)
            for path in ['/health', '/health/auth', '/health/run', '/health/submission']:
                deadline = time.monotonic() + 30
                while True:
                    try:
                        assert request(gateway, path)['status'] == 'success'
                        break
                    except (urllib.error.URLError, AssertionError):
                        if time.monotonic() >= deadline:
                            raise
                        time.sleep(0.25)
            assert redis('PING') == 'PONG'
            assert sql('SELECT count(*) FROM users;') == '0', 'Test DB is not fresh'
            assert 'submission-worker started' in run('logs', '--no-color', 'submission_worker')
            print('All services healthy; fresh schema ready; Redis responds', flush=True)

            # Install the seed schema's exact runtime, never execute source on the host.
            request(piston, '/api/v2/packages', {'language': 'python', 'version': '3.10.0'}, timeout=300)
            assert any(x['language'] == 'python' and x['version'] == '3.10.0'
                       for x in request(piston, '/api/v2/runtimes'))
            print('Piston Python 3.10.0 installed', flush=True)
            cookies = http.cookiejar.CookieJar()
            client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookies))
            user = {'email': 'smoke@example.org', 'username': 'smoke',
                    'fullName': 'Smoke User', 'password': 'SmokeOnly123'}
            registered = request(gateway, '/auth/register', user, expected=201)
            token = registered['data']['accessToken']
            uid = registered['data']['user']['id']
            uuid.UUID(uid)
            request(gateway, '/auth/login', {k: user[k] for k in ['email', 'password']}, opener=client)
            token = request(gateway, '/auth/refresh', {}, opener=client)['data']['accessToken']
            assert request(gateway, '/auth/me', token=token)['data']['user']['id'] == uid
            # Readiness fixture uses direct SQL in this disposable database only.
            sql((ROOT / 'tests/docker/fixture.sql').read_text())
            languages = request(gateway, '/api/v1/languages')['data']['items']
            language = next(x['id'] for x in languages if x['slug'] == 'python')
            problem = request(gateway, '/api/v1/problems/smoke-sum', token=token)['data']
            assert '10 20' not in json.dumps(problem) and '-3 8' not in json.dumps(problem)
            pid = problem['problem']['id']
            code = 'import sys\nprint(sum(map(int, sys.stdin.read().split())))'
            payload = {'problemId': pid, 'languageId': language, 'code': code}
            result = request(gateway, '/api/v1/run', payload, token)['data']
            assert result['status'] == 'Accepted' and result['total'] == 1, result
            assert sql('SELECT count(*) FROM submissions;') == '0', 'Run persisted a submission'
            # Stop worker to observe real Pending + Redis queue, then resume it.
            run('stop', 'submission_worker')
            submitted = request(gateway, '/api/v1/submit', payload, token, expected=201)['data']['submission']
            sid = submitted['id']
            assert submitted['status'] == 'Pending', submitted
            assert redis('TYPE', 'submission_jobs') == 'list'
            assert sid in redis('LRANGE', 'submission_jobs', '0', '-1')
            run('start', 'submission_worker')
            deadline = time.monotonic() + 45
            while time.monotonic() < deadline:
                final = request(gateway, '/submissions/' + sid, token=token)['data']['submission']
                if final['status'] not in ['Pending', 'Running']:
                    break
                time.sleep(0.25)
            assert final['status'] == 'Accepted' and final['passCount'] == 2, final
            assert sql(f"SELECT status FROM submissions WHERE id='{sid}';") == 'Accepted'
            other = request(gateway, '/auth/register', {**user, 'email': 'other@example.org', 'username': 'other'}, expected=201)
            request(gateway, '/submissions/' + sid, token=other['data']['accessToken'], expected=404)
            wrong_token = other['data']['accessToken']
            wrong = request(gateway, '/api/v1/submit', {**payload, 'code': 'print(0)'}, wrong_token, expected=201)['data']['submission']
            deadline = time.monotonic() + 45
            while time.monotonic() < deadline:
                wrong_final = request(gateway, '/submissions/' + wrong['id'], token=wrong_token)['data']['submission']
                if wrong_final['status'] not in ['Pending', 'Running']:
                    break
                time.sleep(0.25)
            assert wrong_final['status'] == 'Wrong Answer', wrong_final
            print('Gateway auth/refresh/catalog, sample Run, Pending -> Redis list -> worker -> Piston -> persisted Accepted: passed', flush=True)

            # These trusted probes verify effective limits of this Piston image.
            def execute(source, timeout=1000):
                return request(piston, '/api/v2/execute', {'language': 'python', 'version': '3.10.0',
                    'files': [{'name': 'main.py', 'content': source}], 'run_timeout': timeout,
                    'run_memory_limit': 134217728})['run']
            assert execute('print(2 + 3)')['stdout'].strip() == '5'
            timed = execute('while True: pass', 500)
            assert timed.get('signal') == 'SIGKILL', timed
            memory = execute("x = bytearray(256 * 1024 * 1024)")
            assert memory.get('code') in [1, 137] or memory.get('signal'), memory
            network = execute("import socket\ns=socket.socket()\ns.settimeout(1)\ntry:\n s.connect(('1.1.1.1', 443))\n print('CONNECTED')\nexcept OSError:\n print('BLOCKED')")
            assert network['stdout'].strip() == 'BLOCKED', network
            output = execute("print('x' * 100000)")
            assert len(output['stdout']) <= 2048, 'Piston output not bounded'
            print('Real Piston timeout/memory/network/output probes: passed', flush=True)
            print('Docker runtime baseline PASSED (not production certification)', flush=True)
        finally:
            print('Removing owned smoke stack and volumes', flush=True)
            run('down', '-v', '--remove-orphans')


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as err:
        # Command output may contain runtime data; show only bounded setup diagnostics.
        raise SystemExit('Docker command failed: ' + err.stderr[-2000:]) from err
