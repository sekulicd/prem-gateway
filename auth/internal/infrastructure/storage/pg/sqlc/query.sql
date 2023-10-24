-- name: InsertRateLimitAndReturnID :one
INSERT INTO rate_limit (requests_per_range, range_in_seconds) VALUES ($1, $2) RETURNING id;

-- name: InsertApiKey :exec
INSERT INTO api_key (id, is_root, rate_limit_id, service_name) VALUES ($1, $2, $3, $4);

-- name: GetAllApiKeys :many
SELECT a.id, a.is_root, a.service_name, r.requests_per_range, r.range_in_seconds
FROM api_key a
    left join rate_limit r on a.rate_limit_id = r.id ORDER BY a.id;

-- name: GetApiKeyForServiceName :many
SELECT a.id, a.is_root, a.service_name, r.requests_per_range, r.range_in_seconds
FROM api_key a
    inner join rate_limit r on a.rate_limit_id = r.id
WHERE a.service_name = $1 ORDER BY a.id;

-- name: GetRootApiKey :one
SELECT id FROM api_key WHERE is_root = true;