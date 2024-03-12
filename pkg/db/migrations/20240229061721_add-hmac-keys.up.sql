SET statement_timeout = 0;

--bun:split

CREATE TABLE subscription_hmac_keys (
    subscription_id INTEGER NOT NULL,
    thirty_day_periods_since_epoch INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    key BYTEA NOT NULL,
    PRIMARY KEY (subscription_id, thirty_day_periods_since_epoch),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

--bun:split

ALTER TABLE subscriptions ADD COLUMN is_silent BOOLEAN DEFAULT FALSE;

--bun:split

-- Ensure that no duplicate subscription rows exist before adding unique index
DELETE FROM subscriptions
WHERE id NOT IN (
    SELECT MAX(id)
    FROM subscriptions
    GROUP BY installation_id, topic
);

--bun:split

CREATE UNIQUE INDEX CONCURRENTLY subscriptions_installation_id_topic_idx ON subscriptions (installation_id, topic);
