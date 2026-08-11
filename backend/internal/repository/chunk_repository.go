package repository

import (
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

type ChunkRepository struct {
	DB *sql.DB
}

func NewChunkRepository(db *sql.DB) *ChunkRepository {
	return &ChunkRepository{
		DB: db,
	}
}

type DocumentChunk struct {
	ID             int
	DocumentID     int
	ChunkIndex     int
	Content        string
	CharacterCount int
}

type CreateChunkInput struct {
	ChunkIndex     int
	Content        string
	CharacterCount int
}

type ChunkSearchResult struct {
	ID             int
	DocumentID     int
	Filename       string
	ChunkIndex     int
	Content        string
	CharacterCount int
	Distance       float64
}

func (r *ChunkRepository) CreateBatch(
	documentID int,
	chunks []CreateChunkInput,
) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin chunk transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO document_chunks (
			document_id,
			chunk_index,
			content,
			character_count
		)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return fmt.Errorf("prepare chunk insert: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		_, err := stmt.Exec(
			documentID,
			chunk.ChunkIndex,
			chunk.Content,
			chunk.CharacterCount,
		)
		if err != nil {
			return fmt.Errorf(
				"insert chunk %d: %w",
				chunk.ChunkIndex,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit chunk transaction: %w", err)
	}

	return nil
}

func (r *ChunkRepository) ListByDocumentID(
	documentID int,
) ([]DocumentChunk, error) {
	rows, err := r.DB.Query(`
		SELECT
			id,
			document_id,
			chunk_index,
			content,
			character_count
		FROM document_chunks
		WHERE document_id = $1
		ORDER BY chunk_index
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []DocumentChunk

	for rows.Next() {
		var chunk DocumentChunk

		err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.CharacterCount,
		)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chunks, nil
}
func (r *DocumentRepository) Delete(id int) error {
	result, err := r.DB.Exec(`
		DELETE FROM documents
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ChunkRepository) ListWithoutEmbedding() ([]DocumentChunk, error) {
	rows, err := r.DB.Query(`
		SELECT
			id,
			document_id,
			chunk_index,
			content,
			character_count
		FROM document_chunks
		WHERE embedding IS NULL
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]DocumentChunk, 0)

	for rows.Next() {
		var chunk DocumentChunk

		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.CharacterCount,
		); err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chunks, nil
}

func (r *ChunkRepository) UpdateEmbedding(
	id int,
	embedding []float32,
) error {
	_, err := r.DB.Exec(`
		UPDATE document_chunks
		SET embedding = $1
		WHERE id = $2
	`,
		pgvector.NewVector(embedding),
		id,
	)

	return err
}

func (r *ChunkRepository) SearchSimilar(
	embedding []float32,
	limit int,
) ([]ChunkSearchResult, error) {
	rows, err := r.DB.Query(`
		SELECT
			dc.id,
			dc.document_id,
			d.filename,
			dc.chunk_index,
			dc.content,
			dc.character_count,
			dc.embedding <=> $1 AS distance
		FROM document_chunks dc
		JOIN documents d
			ON d.id = dc.document_id
		WHERE dc.embedding IS NOT NULL
		ORDER BY dc.embedding <=> $1
		LIMIT $2
	`,
		pgvector.NewVector(embedding),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ChunkSearchResult, 0)

	for rows.Next() {
		var result ChunkSearchResult

		if err := rows.Scan(
			&result.ID,
			&result.DocumentID,
			&result.Filename,
			&result.ChunkIndex,
			&result.Content,
			&result.CharacterCount,
			&result.Distance,
		); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ChunkRepository) ListWithoutEmbeddingByDocumentID(
	documentID int,
) ([]DocumentChunk, error) {
	rows, err := r.DB.Query(`
		SELECT
			id,
			document_id,
			chunk_index,
			content,
			character_count
		FROM document_chunks
		WHERE document_id = $1
		  AND embedding IS NULL
		ORDER BY chunk_index
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]DocumentChunk, 0)

	for rows.Next() {
		var chunk DocumentChunk

		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.CharacterCount,
		); err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chunks, nil
}
