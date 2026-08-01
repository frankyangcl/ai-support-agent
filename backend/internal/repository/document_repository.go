package repository

import "database/sql"

type Document struct {
	ID        int    `json:"id"`
	Filename  string `json:"filename"`
	CreatedAt string `json:"created_at"`
}

type DocumentDetail struct {
	ID        int
	Filename  string
	Content   string
	CreatedAt string
}

type DocumentRepository struct {
	DB *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{
		DB: db,
	}
}

func (r *DocumentRepository) Create(filename, content string) (int, error) {
	var id int

	err := r.DB.QueryRow(
		`INSERT INTO documents (filename, content)
		 VALUES ($1, $2)
		 RETURNING id`,
		filename,
		content,
	).Scan(&id)

	return id, err
}

func (r *DocumentRepository) List() ([]Document, error) {
	rows, err := r.DB.Query(
		`SELECT id, filename, created_at
		 FROM documents
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make([]Document, 0)
	for rows.Next() {
		var doc Document

		if err := rows.Scan(&doc.ID, &doc.Filename, &doc.CreatedAt); err != nil {
			return nil, err
		}

		documents = append(documents, doc)
	}

	return documents, rows.Err()
}

func (r *DocumentRepository) GetByID(id int) (*DocumentDetail, error) {
	var doc DocumentDetail

	err := r.DB.QueryRow(`
		SELECT
			id,
			filename,
			content,
			created_at
		FROM documents
		WHERE id = $1
	`, id).Scan(
		&doc.ID,
		&doc.Filename,
		&doc.Content,
		&doc.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &doc, nil
}
