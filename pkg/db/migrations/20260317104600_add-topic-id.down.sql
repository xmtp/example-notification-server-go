ALTER TABLE subscriptions DROP COLUMN IF EXISTS topic_id;

--bun:split

DROP FUNCTION IF EXISTS populate_topic_ids;

--bun:split

DROP FUNCTION IF EXISTS disable_unsupported_subscriptions;

--bun:split

DROP INDEX IF EXISTS subscriptions_topic_id_is_active_idx;

--bun:split

DROP INDEX IF EXISTS subscriptions_installation_id_topic_id_idx;

--bun:split

-- Recreate previous index (using plain topic).
CREATE UNIQUE INDEX CONCURRENTLY subscriptions_installation_id_topic_idx ON subscriptions (installation_id, topic);

--bun:split

-- Recreate previous index (using plain topic).
CREATE INDEX CONCURRENTLY subscriptions_topic_is_active_idx ON public.subscriptions (topic, is_active);
