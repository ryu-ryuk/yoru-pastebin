package paste

import (
	"time"
)

// represents a single paste entry in the database.
type Paste struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Language     string     `json:"language,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	PasswordHash *string    `json:"-"`
	Salt         []byte     `json:"-"` 
	EncryptedIV  []byte     `json:"-"`
}

// checks if the paste has expired.
func (p *Paste) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false // Never expires
	}
	return time.Now().After(*p.ExpiresAt)
}

// checks if the paste is password protected.
func (p *Paste) IsProtected() bool {
	return p.PasswordHash != nil && *p.PasswordHash != ""
}