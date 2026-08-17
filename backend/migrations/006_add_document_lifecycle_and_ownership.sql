ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS owner_sub TEXT,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ready',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS storage_name TEXT,
    ADD COLUMN IF NOT EXISTS file_size BIGINT,
    ADD COLUMN IF NOT EXISTS mime_type TEXT,
    ADD COLUMN IF NOT EXISTS processing_error TEXT;

ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_status_check;
ALTER TABLE documents ADD CONSTRAINT documents_status_check CHECK (status IN ('processing', 'ready', 'failed'));

DROP INDEX IF EXISTS uq_documents_file_hash;
CREATE UNIQUE INDEX IF NOT EXISTS uq_documents_owner_file_hash ON documents(owner_sub, file_hash)
    WHERE owner_sub IS NOT NULL AND file_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_documents_owner_created ON documents(owner_sub, created_at DESC)
    WHERE owner_sub IS NOT NULL;
