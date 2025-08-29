-- migrate:up
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    payload JSON,
    state VARCHAR(16) NOT NULL CHECK (state IN ('pending', 'started', 'finished', 'failed')),
    error_msg TEXT,
    attempts INT NULL,
    available_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL
);

-- migrate:down
DROP TABLE IF EXISTS jobs;
