-- Recreate the redundant index
CREATE INDEX IF NOT EXISTS device_delivery_mechanisms_installation_id_idx
ON device_delivery_mechanisms (installation_id);

-- Drop covering index
DROP INDEX IF EXISTS device_delivery_mechanisms_latest_idx;
