package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrGenerationConflict = errors.New("generation already active")

type Conversation struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Message struct {
	ID             int64           `json:"id"`
	ConversationID int64           `json:"conversation_id"`
	Role           string          `json:"role"`
	Content        string          `json:"content"`
	Status         string          `json:"status"`
	Citations      json.RawMessage `json:"citations"`
	CreatedAt      time.Time       `json:"created_at"`
}
type Turn struct {
	ConversationID     int64
	UserMessageID      int64
	AssistantMessageID int64
	History            []Message
}
type ConversationRepository struct{ DB *sql.DB }

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{DB: db}
}
func (r *ConversationRepository) RecoverProcessing(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE messages SET status='cancelled' WHERE role='assistant' AND status='processing'`)
	return err
}

func (r *ConversationRepository) Create(ctx context.Context, owner, title string) (*Conversation, error) {
	var c Conversation
	err := r.DB.QueryRowContext(ctx, `INSERT INTO conversations(owner_sub,title) VALUES($1,$2) RETURNING id,title,created_at,updated_at`, owner, title).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (r *ConversationRepository) List(ctx context.Context, owner string) ([]Conversation, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,title,created_at,updated_at FROM conversations WHERE owner_sub=$1 ORDER BY updated_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Conversation, 0)
	for rows.Next() {
		var c Conversation
		if err = rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *ConversationRepository) Get(ctx context.Context, owner string, id int64) (*Conversation, error) {
	var c Conversation
	err := r.DB.QueryRowContext(ctx, `SELECT id,title,created_at,updated_at FROM conversations WHERE id=$1 AND owner_sub=$2`, id, owner).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (r *ConversationRepository) Rename(ctx context.Context, owner string, id int64, title string) (*Conversation, error) {
	var c Conversation
	err := r.DB.QueryRowContext(ctx, `UPDATE conversations SET title=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND owner_sub=$2 RETURNING id,title,created_at,updated_at`, id, owner, title).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (r *ConversationRepository) Delete(ctx context.Context, owner string, id int64) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM conversations WHERE id=$1 AND owner_sub=$2`, id, owner)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func (r *ConversationRepository) Messages(ctx context.Context, owner string, id int64, limit int) ([]Message, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT m.id,m.conversation_id,m.role,m.content,m.status,m.citations,m.created_at FROM messages m JOIN conversations c ON c.id=m.conversation_id WHERE c.id=$1 AND c.owner_sub=$2 ORDER BY m.created_at,m.id LIMIT $3`, id, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Status, &m.Citations, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ConversationRepository) BeginTurn(ctx context.Context, owner string, conversationID *int64, question string, historyLimit int) (*Turn, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var id int64
	if conversationID == nil {
		idTitle := titleFromQuestion(question)
		err = tx.QueryRowContext(ctx, `INSERT INTO conversations(owner_sub,title) VALUES($1,$2) RETURNING id`, owner, idTitle).Scan(&id)
	} else {
		id = *conversationID
		err = tx.QueryRowContext(ctx, `SELECT id FROM conversations WHERE id=$1 AND owner_sub=$2 FOR UPDATE`, id, owner).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	history, err := loadRecentMessages(ctx, tx, id, historyLimit)
	if err != nil {
		return nil, err
	}
	var userID, assistantID int64
	if err = tx.QueryRowContext(ctx, `INSERT INTO messages(conversation_id,role,content,status) VALUES($1,'user',$2,'completed') RETURNING id`, id, question).Scan(&userID); err != nil {
		return nil, err
	}
	if err = tx.QueryRowContext(ctx, `INSERT INTO messages(conversation_id,role,status) VALUES($1,'assistant','processing') RETURNING id`, id).Scan(&assistantID); err != nil {
		if strings.Contains(err.Error(), "uq_messages_active_assistant") {
			return nil, ErrGenerationConflict
		}
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE conversations SET updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Turn{id, userID, assistantID, history}, nil
}
func loadRecentMessages(ctx context.Context, tx *sql.Tx, id int64, limit int) ([]Message, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,conversation_id,role,content,status,citations,created_at FROM (SELECT * FROM messages WHERE conversation_id=$1 AND status='completed' ORDER BY created_at DESC,id DESC LIMIT $2) recent ORDER BY created_at,id`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Status, &m.Citations, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *ConversationRepository) FinishAssistant(ctx context.Context, id int64, status, content string, citations any) error {
	data, err := json.Marshal(citations)
	if err != nil {
		return err
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var conversationID int64
	if err = tx.QueryRowContext(ctx, `UPDATE messages SET status=$2,content=$3,citations=$4 WHERE id=$1 AND role='assistant' AND status='processing' RETURNING conversation_id`, id, status, content, data).Scan(&conversationID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE conversations SET updated_at=CURRENT_TIMESTAMP WHERE id=$1`, conversationID); err != nil {
		return err
	}
	return tx.Commit()
}
func titleFromQuestion(q string) string {
	q = strings.TrimSpace(q)
	r := []rune(q)
	if len(r) > 80 {
		r = r[:80]
	}
	if len(r) == 0 {
		return "New conversation"
	}
	return string(r)
}
