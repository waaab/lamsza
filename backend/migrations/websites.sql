ALTER TABLE users ADD COLUMN IF NOT EXISTS website_banned BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS websites (
    id SERIAL PRIMARY KEY,
    domain_key VARCHAR(253) NOT NULL UNIQUE,
    submitted_host VARCHAR(253) NOT NULL,
    title VARCHAR(120) NOT NULL DEFAULT '',
    description VARCHAR(300) NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL,
    user_id INTEGER REFERENCES users(id),
    entry_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT websites_status_check CHECK (status IN ('pending', 'approved'))
);
CREATE INDEX IF NOT EXISTS idx_websites_status ON websites (status);
CREATE INDEX IF NOT EXISTS idx_websites_entry ON websites (entry_id);
