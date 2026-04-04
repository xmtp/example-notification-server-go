-- Step 1: Drop indexes that reference the topic column
DROP INDEX IF EXISTS subscriptions_topic_is_active_idx;
DROP INDEX IF EXISTS subscriptions_installation_id_topic_idx;

-- Step 2: Rename existing column
ALTER TABLE subscriptions RENAME COLUMN topic TO topic_legacy;

-- Step 3: Add new binary column
ALTER TABLE subscriptions ADD COLUMN topic BYTEA;

-- Step 4: Single-pass conversion using lower() inline (topic_legacy is not modified)
UPDATE subscriptions
SET topic = CASE
    WHEN lower(topic_legacy) ~ '^/xmtp/mls/1/g-([0-9a-f]{2})+/proto$'
    THEN decode('00', 'hex') || decode(
        substring(lower(topic_legacy) FROM '/xmtp/mls/1/g-(.+)/proto$'), 'hex')
    WHEN lower(topic_legacy) ~ '^/xmtp/mls/1/w-([0-9a-f]{2})+/proto$'
    THEN decode('01', 'hex') || decode(
        substring(lower(topic_legacy) FROM '/xmtp/mls/1/w-(.+)/proto$'), 'hex')
    ELSE NULL
END;

-- Step 5: Delete rows that couldn't be converted
DELETE FROM subscriptions WHERE topic IS NULL;

-- Step 6: Deduplicate rows that collapsed to the same (installation_id, topic) after
-- hex normalization (e.g., rows differing only by hex casing). Keep the newest row.
DELETE FROM subscriptions older
USING subscriptions newer
WHERE older.installation_id = newer.installation_id
  AND older.topic = newer.topic
  AND older.id < newer.id;

-- Step 7: Add NOT NULL constraint
ALTER TABLE subscriptions ALTER COLUMN topic SET NOT NULL;

-- Step 8: Recreate indexes on new column
CREATE INDEX subscriptions_topic_is_active_idx
    ON subscriptions (topic, is_active);

CREATE UNIQUE INDEX subscriptions_installation_id_topic_idx
    ON subscriptions (installation_id, topic);

-- Step 9: Drop legacy column
ALTER TABLE subscriptions DROP COLUMN topic_legacy;
