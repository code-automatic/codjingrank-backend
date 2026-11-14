#!/usr/bin/env python3
"""Validate captured real HTTP responses against approved OpenAPI schemas.

Usage: python3 tests/sprint2_contracts.py /tmp/responses.jsonl [docs/openapi_fe.yaml]
Requires PyYAML and jsonschema (documentation/test tooling only).
"""
import copy
import json
import sys
from pathlib import Path
import yaml
from jsonschema import Draft7Validator, FormatChecker

root = Path(__file__).resolve().parents[1]
spec = yaml.safe_load((root / (sys.argv[2] if len(sys.argv) > 2 else 'docs/openapi.yaml')).read_text())

def adapt(value):
    if isinstance(value, list):
        return [adapt(v) for v in value]
    if not isinstance(value, dict):
        return value
    value = {k: adapt(v) for k, v in value.items()}
    if value.pop('nullable', False):
        value = {'anyOf': [value, {'type': 'null'}]}
    return value

spec = adapt(spec)
count = 0
for line in Path(sys.argv[1]).read_text().splitlines():
    record = json.loads(line)
    schema = copy.deepcopy(spec['paths'][record['pattern']][record['method'].lower()]['responses'][str(record['status'])]['content']['application/json']['schema'])
    schema['components'] = spec['components']
    errors = list(Draft7Validator(schema, format_checker=FormatChecker()).iter_errors(record['body']))
    if errors:
        raise SystemExit(f"{record['method']} {record['pattern']}: " + '; '.join(f'{list(e.path)}: {e.message}' for e in errors))
    count += 1
if count == 0:
    raise SystemExit('No responses captured')
print(f'Validated {count} real HTTP responses against OpenAPI')
