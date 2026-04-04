-- Covering index for LATERAL latest-installation lookup
CREATE INDEX IF NOT EXISTS device_delivery_mechanisms_latest_idx
ON device_delivery_mechanisms (installation_id, updated_at DESC, id DESC)
INCLUDE (kind, token, created_at);

-- Drop redundant prefix index (subsumed by unique constraint on (installation_id, kind, token))
DROP INDEX IF EXISTS device_delivery_mechanisms_installation_id_idx;

-- One-time cleanup: deactivate subscriptions for already-deleted installations
UPDATE subscriptions s
SET is_active = FALSE
FROM installations i
WHERE s.installation_id = i.id
  AND i.deleted_at IS NOT NULL
  AND s.is_active = TRUE;
