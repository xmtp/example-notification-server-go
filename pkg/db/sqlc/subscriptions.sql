-- name: ReactivateSubscriptions :many
UPDATE subscriptions
SET is_active = TRUE
WHERE installation_id = sqlc.arg(installation_id)
  AND topic = ANY(sqlc.arg(topics)::bytea[])
RETURNING id, created_at, installation_id, topic, is_active, is_silent;

-- name: DeactivateSubscriptions :exec
UPDATE subscriptions
SET is_active = FALSE
WHERE installation_id = sqlc.arg(installation_id)
  AND topic = ANY(sqlc.arg(topics)::bytea[]);

-- name: ListActiveSubscriptionsByTopicAndPeriod :many
SELECT
    s.id,
    s.created_at,
    s.installation_id,
    s.topic,
    s.is_active,
    s.is_silent,
    shk.subscription_id IS NOT NULL AS has_hmac_key,
    shk.thirty_day_periods_since_epoch,
    shk.key
FROM subscriptions AS s
LEFT JOIN subscription_hmac_keys AS shk
    ON shk.subscription_id = s.id
   AND shk.thirty_day_periods_since_epoch = sqlc.arg(thirty_day_period)
WHERE s.topic = sqlc.arg(topic)
  AND s.is_active = TRUE
ORDER BY s.id;

-- name: BatchUpsertSubscriptions :many
INSERT INTO subscriptions (installation_id, topic, is_active, is_silent)
SELECT sqlc.arg(installation_id)::text, t.topic, TRUE, t.is_silent
FROM ROWS FROM (
    unnest(sqlc.arg(topics)::bytea[]),
    unnest(sqlc.arg(is_silents)::boolean[])
) AS t(topic, is_silent)
ON CONFLICT (installation_id, topic) DO UPDATE
SET is_active = TRUE, is_silent = EXCLUDED.is_silent
RETURNING id, topic;

-- name: BatchUpsertSubscriptionHmacKeys :exec
INSERT INTO subscription_hmac_keys (subscription_id, thirty_day_periods_since_epoch, key)
SELECT t.sub_id, t.period, t.hmac_key
FROM ROWS FROM (
    unnest(sqlc.arg(subscription_ids)::bigint[]),
    unnest(sqlc.arg(periods)::integer[]),
    unnest(sqlc.arg(keys)::bytea[])
) AS t(sub_id, period, hmac_key)
ON CONFLICT (subscription_id, thirty_day_periods_since_epoch) DO UPDATE
SET key = EXCLUDED.key, updated_at = NOW();

-- name: BatchInsertSubscriptions :exec
INSERT INTO subscriptions (installation_id, topic, is_active, is_silent)
SELECT sqlc.arg(installation_id)::text, t.topic, TRUE, FALSE
FROM unnest(sqlc.arg(topics)::bytea[]) AS t(topic)
ON CONFLICT (installation_id, topic) DO NOTHING;

-- name: DeactivateInstallationSubscriptions :exec
UPDATE subscriptions
SET is_active = FALSE
WHERE installation_id = sqlc.arg(installation_id)
  AND is_active = TRUE;
