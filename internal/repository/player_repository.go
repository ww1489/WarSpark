package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type PlayerRepository struct {
	db *sqlx.DB
}

func NewPlayerRepository(db *sqlx.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

type SavePlayerSnapshotInput struct {
	ID        string
	PlayerTag string
	Overview  clandomain.PlayerOverview
	FetchedAt string
}

func (r *PlayerRepository) SaveSnapshot(ctx context.Context, input SavePlayerSnapshotInput) error {
	overviewJSON, err := json.Marshal(input.Overview)
	if err != nil {
		return fmt.Errorf("marshal player overview: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO player_snapshots (id, player_tag, overview, fetched_at) VALUES (?, ?, ?, ?)",
		input.ID, input.PlayerTag, string(overviewJSON), input.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("insert player snapshot: %w", err)
	}
	return nil
}
