package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/frankyangcl/ai-support-agent/backend/internal/chunker"
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

var ErrDuplicateDocument = errors.New("document already exists")
var ErrInvalidState = errors.New("invalid document state")

type DocumentService struct {
	Repo             *repository.DocumentRepository
	ChunkRepo        *repository.ChunkRepository
	PDFParser        *parser.PDFParser
	TextChunker      *chunker.TextChunker
	EmbeddingService *EmbeddingService
	UploadDir        string
}

func NewDocumentService(repo *repository.DocumentRepository, chunkRepo *repository.ChunkRepository, pdf *parser.PDFParser, text *chunker.TextChunker, embedding *EmbeddingService) *DocumentService {
	return &DocumentService{repo, chunkRepo, pdf, text, embedding, "uploads"}
}

func (s *DocumentService) CreateDocument(owner, filename, content string) (int, error) {
	return s.Repo.Create(owner, filename, content, "")
}
func (s *DocumentService) ListDocuments(owner string) ([]repository.Document, error) {
	return s.Repo.List(owner)
}
func (s *DocumentService) GetDocument(owner string, id int) (*repository.DocumentDetail, error) {
	return s.Repo.GetByID(owner, id)
}

func (s *DocumentService) CreateDocumentFromPDF(ctx context.Context, owner, filename, path, storage, mime string, size int64) (*repository.DocumentDetail, error) {
	hash, err := calculateFileHash(path)
	if err != nil {
		return nil, err
	}
	exists, err := s.Repo.ExistsByHash(owner, hash)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, ErrDuplicateDocument
	}
	id, err := s.Repo.CreateProcessing(owner, filename, storage, hash, size, mime)
	if err != nil {
		return nil, fmt.Errorf("create processing document: %w", err)
	}
	processErr := s.process(ctx, owner, id, path)
	if processErr != nil {
		_ = s.Repo.MarkFailed(owner, id, "processing_failed")
		log.Printf("document processing failed document_id=%d category=processing_failed: %v", id, processErr)
	} else {
		log.Printf("document processing ready document_id=%d", id)
	}
	doc, getErr := s.Repo.GetByID(owner, id)
	if getErr != nil {
		return nil, getErr
	}
	return doc, processErr
}

func (s *DocumentService) process(ctx context.Context, owner string, id int, path string) error {
	text, err := s.PDFParser.ExtractText(path)
	if err != nil {
		return fmt.Errorf("extract PDF: %w", err)
	}
	chunks := s.TextChunker.Split(text)
	if len(chunks) == 0 {
		return errors.New("PDF produced no text chunks")
	}
	if err = s.ChunkRepo.DeleteByDocumentID(id); err != nil {
		return fmt.Errorf("clear chunks: %w", err)
	}
	inputs := make([]repository.CreateChunkInput, 0, len(chunks))
	for _, c := range chunks {
		inputs = append(inputs, repository.CreateChunkInput{ChunkIndex: c.Index, Content: c.Content, CharacterCount: c.CharacterCount})
	}
	if err = s.ChunkRepo.CreateBatch(id, inputs); err != nil {
		return fmt.Errorf("create chunks: %w", err)
	}
	count, err := s.EmbeddingService.ProcessDocumentChunks(ctx, id)
	if err != nil {
		return fmt.Errorf("embed chunks: %w", err)
	}
	if count != len(chunks) {
		return fmt.Errorf("embedding count mismatch")
	}
	if err = s.Repo.MarkReady(owner, id, text); err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}
	return nil
}

func (s *DocumentService) Retry(ctx context.Context, owner string, id int) (*repository.DocumentDetail, error) {
	doc, err := s.Repo.BeginRetry(owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		if _, getErr := s.Repo.GetByID(owner, id); getErr != nil {
			return nil, sql.ErrNoRows
		}
		return nil, ErrInvalidState
	}
	if err != nil {
		return nil, err
	}
	if doc.StorageName == "" {
		_ = s.Repo.MarkFailed(owner, id, "source_unavailable")
		return nil, errors.New("source file unavailable")
	}
	err = s.process(ctx, owner, id, filepath.Join(s.UploadDir, doc.StorageName))
	if err != nil {
		_ = s.Repo.MarkFailed(owner, id, "processing_failed")
		log.Printf("document retry failed document_id=%d category=processing_failed: %v", id, err)
	}
	updated, getErr := s.Repo.GetByID(owner, id)
	if getErr != nil {
		return nil, getErr
	}
	return updated, err
}

func (s *DocumentService) DeleteDocument(owner string, id int) error {
	doc, err := s.Repo.GetByID(owner, id)
	if err != nil {
		return err
	}
	if doc.Status == "processing" {
		return ErrInvalidState
	}
	storage, err := s.Repo.Delete(owner, id)
	if err != nil {
		return err
	}
	if storage != "" {
		path := filepath.Join(s.UploadDir, filepath.Base(storage))
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Printf("document file cleanup failed document_id=%d category=storage_cleanup", id)
		}
	}
	return nil
}

func calculateFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
