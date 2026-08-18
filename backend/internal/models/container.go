package models

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lovelymondayz/say-less/backend/internal/spotify"
)

// Container holds all dependencies
type Container struct {
	DB          *pgxpool.Pool
	Spotify     *spotify.Client
}
