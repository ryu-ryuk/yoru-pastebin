package paste

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"        // For ErrNoRows
	"github.com/jackc/pgx/v5/pgxpool" // For pgxpool.Pool
)

// defines the interface for paste data operations.
type Repository interface {
	CreatePaste(ctx context.Context, paste *Paste) error
	GetPasteByID(ctx context.Context, id string) (*Paste, error)
	DeletePasteByID(ctx context.Context, id string) error
	// TODO: Add methods for listing/searching pastes // maybe
}

type PGRepository struct {
	pool *pgxpool.Pool 
}

// creates a new PostgreSQL paste repository.
func NewPGRepository(pool *pgxpool.Pool) Repository { 
	return &PGRepository{pool: pool}
}

// inserts a new paste into the database.
func (r *PGRepository) CreatePaste(ctx context.Context, paste *Paste) error {
	query := `
        INSERT INTO pastes (id, content, created_at, expires_at, password_hash, salt, encrypted_iv, language)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.pool.Exec(
		ctx,
		query,
		paste.ID,
		paste.Content,
		paste.CreatedAt,
		paste.ExpiresAt,
		paste.PasswordHash,
		paste.Salt,
		paste.EncryptedIV,
		paste.Language,
	)
	if err != nil {
		return fmt.Errorf("failed to create paste: %w", err)
	}
	return nil
}

// this retrieves a paste by its ID.
func (r *PGRepository) GetPasteByID(ctx context.Context, id string) (*Paste, error) {
	paste := &Paste{}
	query := `
        SELECT id, content, created_at, expires_at, password_hash, salt, encrypted_iv, language
        FROM pastes
        WHERE id = $1 AND (expires_at IS NULL OR expires_at > $2)
    `
	err := r.pool.QueryRow(ctx, query, id, time.Now()).Scan(
		&paste.ID,
		&paste.Content,
		&paste.CreatedAt,
		&paste.ExpiresAt,
		&paste.PasswordHash,
		&paste.Salt,
		&paste.EncryptedIV,
		&paste.Language,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("paste not found")
		}
		return nil, fmt.Errorf("failed to get paste by ID %s: %w", id, err)
	}
	return paste, nil
}

// this removes a paste from the database.
func (r *PGRepository) DeletePasteByID(ctx context.Context, id string) error {
	query := `DELETE FROM pastes WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete paste by ID %s: %w", id, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no paste found with ID %s to delete", id)
	}
	return nil
}