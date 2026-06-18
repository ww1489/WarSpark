package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type WarRepository struct {
	db *sqlx.DB
}

func NewWarRepository(db *sqlx.DB) *WarRepository {
	return &WarRepository{db: db}
}

func (r *WarRepository) SaveSnapshot(ctx context.Context, input wardomain.SaveSnapshotInput) (wardomain.Snapshot, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return wardomain.Snapshot{}, fmt.Errorf("begin save war snapshot: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO war_snapshots (
  id, clan_tag, opponent_clan_tag, war_state, team_size, clan_stars,
  opponent_stars, clan_destruction, opponent_destruction, source_type, fetched_at
) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, 'official_api', ?)`,
		input.ID,
		input.ClanTag,
		input.OpponentClanTag,
		input.WarState,
		nullInt(input.TeamSize),
		input.ClanStars,
		input.OpponentStars,
		input.ClanDestruction,
		input.OpponentDestruction,
		input.FetchedAt,
	)
	if err != nil {
		return wardomain.Snapshot{}, fmt.Errorf("insert war snapshot: %w", err)
	}

	for _, member := range input.Members {
		_, err := tx.ExecContext(ctx, `
INSERT INTO war_members (
  id, war_snapshot_id, side, map_position, player_tag, player_name, th_level,
  attacks_used, best_stars_against, best_destruction_against
) VALUES (?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?)`,
			member.ID,
			input.ID,
			member.Side,
			member.MapPosition,
			member.PlayerTag,
			member.PlayerName,
			member.THLevel,
			member.AttacksUsed,
			member.BestStarsAgainst,
			member.BestDestructionAgainst,
		)
		if err != nil {
			return wardomain.Snapshot{}, fmt.Errorf("insert war member: %w", err)
		}
	}

	for _, target := range input.Targets {
		_, err := tx.ExecContext(ctx, `
INSERT INTO war_targets (
  id, war_snapshot_id, war_member_id, target_position, target_name, target_th
) VALUES (?, ?, ?, ?, ?, ?)`,
			target.ID,
			input.ID,
			target.WarMemberID,
			target.TargetPosition,
			target.TargetName,
			target.TargetTH,
		)
		if err != nil {
			return wardomain.Snapshot{}, fmt.Errorf("insert war target: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return wardomain.Snapshot{}, fmt.Errorf("commit save war snapshot: %w", err)
	}
	return r.snapshot(ctx, input.ID)
}

func (r *WarRepository) ListMembers(ctx context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error) {
	where := "WHERE war_snapshot_id = ?"
	args := []any{snapshotID}
	if side != "" {
		where += " AND side = ?"
		args = append(args, side)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM war_members "+where, args...); err != nil {
		return wardomain.MemberListResult{}, fmt.Errorf("count war members: %w", err)
	}

	var items []wardomain.Member
	queryArgs := append(append([]any{}, args...), pagination.PageSize, pagination.Offset)
	if err := r.db.SelectContext(ctx, &items, `
SELECT
  id,
  war_snapshot_id,
  side,
  map_position,
  COALESCE(player_tag, '') AS player_tag,
  player_name,
  th_level,
  attacks_used,
  best_stars_against,
  best_destruction_against
FROM war_members
`+where+`
ORDER BY side ASC, map_position ASC
LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return wardomain.MemberListResult{}, fmt.Errorf("list war members: %w", err)
	}

	return wardomain.MemberListResult{Items: items, Total: total}, nil
}

func (r *WarRepository) snapshot(ctx context.Context, id string) (wardomain.Snapshot, error) {
	var snapshot wardomain.Snapshot
	if err := r.db.GetContext(ctx, &snapshot, `
SELECT
  id,
  clan_tag,
  COALESCE(opponent_clan_tag, '') AS opponent_clan_tag,
  war_state,
  COALESCE(team_size, 0) AS team_size,
  clan_stars,
  opponent_stars,
  clan_destruction,
  opponent_destruction,
  fetched_at
FROM war_snapshots
WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return wardomain.Snapshot{}, wardomain.ErrSnapshotNotFound
		}
		return wardomain.Snapshot{}, fmt.Errorf("get war snapshot: %w", err)
	}
	return snapshot, nil
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
