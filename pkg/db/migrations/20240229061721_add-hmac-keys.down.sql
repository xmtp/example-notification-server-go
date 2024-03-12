SET statement_timeout = 0;

--bun:split

DROP TABLE subscription_hmac_keys;

--bun:split

ALTER TABLE subscriptions DROP COLUMN is_silent;

--bun:split

DROP INDEX CONCURRENTLY IF EXISTS subscriptions_installation_id_topic_idx;
