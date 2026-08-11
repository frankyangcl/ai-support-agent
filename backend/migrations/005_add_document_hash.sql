ALTER TABLE documents
ADD COLUMN IF NOT EXISTS file_hash TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_documents_file_hash
ON documents(file_hash)
WHERE file_hash IS NOT NULL;