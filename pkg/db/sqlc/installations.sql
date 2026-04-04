-- name: UpsertInstallation :exec
INSERT INTO installations (id)
VALUES (sqlc.arg(id))
ON CONFLICT (id) DO UPDATE
SET deleted_at = NULL;

-- name: UpsertDeviceDeliveryMechanism :exec
INSERT INTO device_delivery_mechanisms (
    installation_id,
    kind,
    token,
    updated_at
)
VALUES (
    sqlc.arg(installation_id),
    sqlc.arg(kind),
    sqlc.arg(token),
    COALESCE(sqlc.narg(updated_at), clock_timestamp())
)
ON CONFLICT (installation_id, kind, token) DO UPDATE
SET updated_at = EXCLUDED.updated_at;

-- name: SoftDeleteInstallation :exec
UPDATE installations
SET deleted_at = sqlc.arg(deleted_at)
WHERE id = sqlc.arg(id);

-- name: GetLatestInstallations :many
SELECT ddm.id, ddm.installation_id, ddm.kind, ddm.token, ddm.created_at, ddm.updated_at
FROM installations i
JOIN LATERAL (
    SELECT d.id, d.installation_id, d.kind, d.token, d.created_at, d.updated_at
    FROM device_delivery_mechanisms d
    WHERE d.installation_id = i.id
    ORDER BY d.updated_at DESC, d.id DESC
    LIMIT 1
) ddm ON TRUE
WHERE i.id = ANY(sqlc.arg(installation_ids)::text[])
  AND i.deleted_at IS NULL
ORDER BY i.id DESC;
