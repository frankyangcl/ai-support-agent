ALTER TABLE document_chunks
DROP COLUMN IF EXISTS embedding;

ALTER TABLE document_chunks
ADD COLUMN embedding vector(1024);