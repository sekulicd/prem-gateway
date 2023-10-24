CREATE TABLE rate_limit (
    id SERIAL PRIMARY KEY,
    requests_per_range INT,
    range_in_seconds INT
);

CREATE TABLE api_key (
    id VARCHAR(255) PRIMARY KEY,
    is_root BOOLEAN,
    rate_limit_id INT,
    service_name VARCHAR(255) UNIQUE,
    FOREIGN KEY (rate_limit_id) REFERENCES rate_limit(id)
);