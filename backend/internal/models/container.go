package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lovelymondayz/say-less/backend/internal/spotify"
)

type Container struct {
	DB      *pgxpool.Pool
	Spotify *spotify.Client
}

type Generation struct {
	ID            string    `json:"id"`
	ShareID       string    `json:"share_id"`
	OriginalText  string    `json:"original_text"`
	Mode          string    `json:"mode"`
	Tracks        []byte    `json:"tracks"`
	Reconstructed string    `json:"reconstructed"`
	Caption       string    `json:"caption"`
	Accuracy      float64   `json:"accuracy"`
	CreatedAt     time.Time `json:"created_at"`
}

func Migrate(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS generations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			share_id VARCHAR(12) UNIQUE NOT NULL,
			original_text TEXT NOT NULL,
			mode VARCHAR(10) NOT NULL,
			tracks JSONB NOT NULL,
			reconstructed TEXT NOT NULL DEFAULT '',
			caption TEXT NOT NULL DEFAULT '',
			accuracy DOUBLE PRECISION DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_generations_share_id ON generations(share_id);
	`)
	return err
}
