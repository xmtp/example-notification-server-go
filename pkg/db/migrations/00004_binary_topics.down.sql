DROP INDEX IF EXISTS subscriptions_topic_is_active_idx;
DROP INDEX IF EXISTS subscriptions_installation_id_topic_idx;

ALTER TABLE subscriptions ADD COLUMN topic_legacy TEXT;
ALTER TABLE subscriptions DROP COLUMN topic;
ALTER TABLE subscriptions RENAME COLUMN topic_legacy TO topic;

CREATE INDEX subscriptions_topic_is_active_idx
    ON subscriptions (topic, is_active);
CREATE UNIQUE INDEX subscriptions_installation_id_topic_idx
    ON subscriptions (installation_id, topic);
