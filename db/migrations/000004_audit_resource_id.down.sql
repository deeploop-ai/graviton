DROP INDEX IF EXISTS idx_audit_logs_resource;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS resource_id;
