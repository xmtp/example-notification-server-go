SET statement_timeout = 0;

--bun:split

ALTER TABLE subscriptions ADD COLUMN topic_id TEXT DEFAULT '';

--bun:split

-- Populate the new column based on the old one.
CREATE OR REPLACE FUNCTION populate_topic_ids()
RETURNS VOID AS $$
BEGIN
    UPDATE subscriptions
    SET topic_id = regexp_replace(topic, '^/xmtp/mls/1/[gw]-', '')
    WHERE topic_id = '';
END;
$$ LANGUAGE plpgsql;

--bun:split

SELECT populate_topic_ids();

--bun:split

-- Turn off all subscriptions which have invalid topic IDs.
CREATE OR REPLACE FUNCTION disable_unsupported_subscriptions()
RETURNS VOID AS $$
BEGIN
    UPDATE subscriptions
    SET is_silent = TRUE
    WHERE topic NOT LIKE '/xmtp/mls/1/g-%' AND topic NOT LIKE '/xmtp/mls/1/w-%';
END;
$$ LANGUAGE plpgsql;

--bun:split

SELECT disable_unsupported_subscriptions();

--bun:split

CREATE INDEX CONCURRENTLY subscriptions_topic_id_is_active_idx ON subscriptions (topic_id, is_active);

--bun:split

DROP INDEX IF EXISTS subscriptions_topic_is_active_idx;

--bun:split

-- Create an updated index using installation_id and topic_id, analogous to the "add-hmac-keys" migration.
-- Ensure that no duplicate subscription rows exist before adding unique index
DELETE FROM subscriptions
WHERE id NOT IN (
    SELECT MAX(id)
    FROM subscriptions
    GROUP BY installation_id, topic_id
);

--bun:split

CREATE UNIQUE INDEX CONCURRENTLY subscriptions_installation_id_topic_id_idx ON subscriptions (installation_id, topic_id);

--bun:split

-- Remove previous index which will not be used anymore.
DROP INDEX IF EXISTS subscriptions_installation_id_topic_idx;
