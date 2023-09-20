/* DNS_INFO */

-- name: InsertDnsInfo :exec
INSERT INTO dns_info(domain, sub_domain, ip, node_name, email) VALUES ($1, $2, $3, $4, $5);

-- name: UpdateDnsInfo :exec
UPDATE dns_info SET sub_domain = $1, ip = $2, node_name = $3, email = $4 WHERE domain = $5;

-- name: DeleteDnsInfo :exec
DELETE FROM dns_info WHERE domain = $1;

-- name: GetDnsInfo :one
SELECT * FROM dns_info WHERE domain = $1;

-- name: GetExistDnsInfo :one
SELECT * FROM dns_info;