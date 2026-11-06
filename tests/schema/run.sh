#!/usr/bin/env bash
set -euo pipefail
# Only a newly created temporary database is touched; existing databases are never reset.
container=${SCHEMA_TEST_CONTAINER:?Set SCHEMA_TEST_CONTAINER to a PostgreSQL 16 container}
user=${SCHEMA_TEST_USER:?Set SCHEMA_TEST_USER to its database admin role}
root=$(cd "$(dirname "$0")/../.." && pwd)
database="schema_test_${RANDOM}_$$"
docker exec "$container" createdb -U "$user" "$database"
trap 'docker exec "$container" dropdb -U "$user" "$database"' EXIT
for pass in 1 2; do
    docker exec -i "$container" psql -X -q -U "$user" -d "$database" -v ON_ERROR_STOP=1 < "$root/schema.sql"
done
docker exec -i "$container" psql -X -q -U "$user" -d "$database" -v ON_ERROR_STOP=1 < "$root/tests/schema/constraints.sql"
