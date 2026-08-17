package repository

import (
	"database/sql"
	"time"
)

type Document struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	FileSize   int64     `json:"file_size,omitempty"`
	MimeType   string    `json:"mime_type,omitempty"`
	ChunkCount int       `json:"chunk_count"`
}
type DocumentDetail struct {
	Document
	StorageName string `json:"-"`
}
type DocumentRepository struct{ DB *sql.DB }

func NewDocumentRepository(db *sql.DB) *DocumentRepository { return &DocumentRepository{DB: db} }

func (r *DocumentRepository) CreateProcessing(owner, filename, storage, hash string, size int64, mime string) (int, error) {
	var id int
	err := r.DB.QueryRow(`INSERT INTO documents(owner_sub,filename,content,file_hash,status,storage_name,file_size,mime_type) VALUES($1,$2,'',$3,'processing',$4,$5,$6) RETURNING id`, owner, filename, hash, storage, size, mime).Scan(&id)
	return id, err
}
func (r *DocumentRepository) Create(owner, filename, content, hash string) (int, error) {
	var id int
	err := r.DB.QueryRow(`INSERT INTO documents(owner_sub,filename,content,file_hash,status) VALUES($1,$2,$3,$4,'ready') RETURNING id`, owner, filename, content, hash).Scan(&id)
	return id, err
}

const documentSelect = `SELECT d.id,d.filename,d.status,d.created_at,d.updated_at,COALESCE(d.file_size,0),COALESCE(d.mime_type,''),COUNT(dc.id),COALESCE(d.storage_name,'') FROM documents d LEFT JOIN document_chunks dc ON dc.document_id=d.id`

func scanDocument(s interface{ Scan(...any) error }) (*DocumentDetail, error) {
	var d DocumentDetail
	err := s.Scan(&d.ID, &d.Filename, &d.Status, &d.CreatedAt, &d.UpdatedAt, &d.FileSize, &d.MimeType, &d.ChunkCount, &d.StorageName)
	return &d, err
}
func (r *DocumentRepository) List(owner string) ([]Document, error) {
	rows, err := r.DB.Query(documentSelect+` WHERE d.owner_sub=$1 GROUP BY d.id ORDER BY d.created_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	docs := make([]Document, 0)
	for rows.Next() {
		d, e := scanDocument(rows)
		if e != nil {
			return nil, e
		}
		docs = append(docs, d.Document)
	}
	return docs, rows.Err()
}
func (r *DocumentRepository) GetByID(owner string, id int) (*DocumentDetail, error) {
	return scanDocument(r.DB.QueryRow(documentSelect+` WHERE d.owner_sub=$1 AND d.id=$2 GROUP BY d.id`, owner, id))
}
func (r *DocumentRepository) ExistsByHash(owner, hash string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM documents WHERE owner_sub=$1 AND file_hash=$2)`, owner, hash).Scan(&exists)
	return exists, err
}
func (r *DocumentRepository) MarkReady(owner string, id int, content string) error {
	res, err := r.DB.Exec(`UPDATE documents SET content=$3,status='ready',processing_error=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND owner_sub=$2 AND status='processing'`, id, owner, content)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func (r *DocumentRepository) MarkFailed(owner string, id int, category string) error {
	_, err := r.DB.Exec(`UPDATE documents SET status='failed',processing_error=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND owner_sub=$2`, id, owner, category)
	return err
}
func (r *DocumentRepository) BeginRetry(owner string, id int) (*DocumentDetail, error) {
	res, err := r.DB.Exec(`UPDATE documents SET status='processing',processing_error=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND owner_sub=$2 AND status='failed'`, id, owner)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return r.GetByID(owner, id)
}
func (r *DocumentRepository) Delete(owner string, id int) (string, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var storage string
	err = tx.QueryRow(`SELECT COALESCE(storage_name,'') FROM documents WHERE id=$1 AND owner_sub=$2 AND status<>'processing' FOR UPDATE`, id, owner).Scan(&storage)
	if err != nil {
		return "", err
	}
	if _, err = tx.Exec(`DELETE FROM documents WHERE id=$1 AND owner_sub=$2`, id, owner); err != nil {
		return "", err
	}
	return storage, tx.Commit()
}
