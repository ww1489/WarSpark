package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type ClanRepository struct {
	db *sqlx.DB
}

func NewClanRepository(db *sqlx.DB) *ClanRepository {
	return &ClanRepository{db: db}
}

type SaveClanSnapshotInput struct {
	ID        string
	ClanTag   string
	Detail    clandomain.ClanDetail
	FetchedAt string
}

func (r *ClanRepository) SaveSnapshot(ctx context.Context, input SaveClanSnapshotInput) error {
	detailJSON, err := json.Marshal(input.Detail)
	if err != nil {
		return fmt.Errorf("marshal clan detail: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO clan_snapshots (id, clan_tag, detail, fetched_at) VALUES (?, ?, ?, ?)",
		input.ID, input.ClanTag, string(detailJSON), input.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("insert clan snapshot: %w", err)
	}
	return nil
}
