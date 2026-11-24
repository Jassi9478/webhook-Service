CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload TEXT NOT NULL,
    idempotency_key VARCHAR(128) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS delivery_attempts (
    id VARCHAR(64) PRIMARY KEY,
    event_id VARCHAR(64) NOT NULL,
    target_url TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    response_code INT,
    attempt_number INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);