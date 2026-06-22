ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS resource_id TEXT;
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_id, created_at DESC) WHERE resource_id IS NOT NULL;
