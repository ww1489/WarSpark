package war

import (
	"errors"
	"time"
)

const (
	ErrorInvalidTag         = "invalid_tag"
	ErrorAPINotConfigured   = "api_not_configured"
	ErrorAPIRequestFailed   = "api_request_failed"
	ErrorAPIAccessDenied    = "api_access_denied"
	ErrorWarNotFound        = "war_not_found"
	ErrorAPIResponseInvalid = "api_response_invalid"
	ErrorSnapshotNotFound   = "snapshot_not_found"
)

var ErrSnapshotNotFound = errors.New("war snapshot not found")

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Err
}

func NewError(code string, message string) Error {
	return Error{Code: code, Message: message}
}

func WrapError(code string, message string, err error) Error {
	return Error{Code: code, Message: message, Err: err}
}

func ErrorCode(err error) string {
	var warErr Error
	if errors.As(err, &warErr) {
		return warErr.Code
	}
	return ""
}

type CurrentWar struct {
	State    string  `json:"state"`
	TeamSize int     `json:"teamSize"`
	Clan     WarClan `json:"clan"`
	Opponent WarClan `json:"opponent"`
}

type WarClan struct {
	Tag                   string      `json:"tag"`
	Name                  string      `json:"name"`
	Stars                 int         `json:"stars"`
	DestructionPercentage float64     `json:"destructionPercentage"`
	Members               []WarMember `json:"members"`
}

type WarMember struct {
	Tag           string      `json:"tag"`
	Name          string      `json:"name"`
	TownHallLevel int         `json:"townhallLevel"`
	MapPosition   int         `json:"mapPosition"`
	Attacks       []WarAttack `json:"attacks"`
}

type WarAttack struct {
	AttackerTag           string  `json:"attackerTag"`
	DefenderTag           string  `json:"defenderTag"`
	Stars                 int     `json:"stars"`
	DestructionPercentage float64 `json:"destructionPercentage"`
	Order                 int     `json:"order"`
	Duration              int     `json:"duration"`
}

type Snapshot struct {
	ID                  string    `json:"war_snapshot_id" db:"id"`
	ClanTag             string    `json:"clan_tag" db:"clan_tag"`
	OpponentClanTag     string    `json:"opponent_clan_tag,omitempty" db:"opponent_clan_tag"`
	WarState            string    `json:"war_state" db:"war_state"`
	TeamSize            int       `json:"team_size,omitempty" db:"team_size"`
	ClanStars           *int      `json:"clan_stars,omitempty" db:"clan_stars"`
	OpponentStars       *int      `json:"opponent_stars,omitempty" db:"opponent_stars"`
	ClanDestruction     *float64  `json:"clan_destruction,omitempty" db:"clan_destruction"`
	OpponentDestruction *float64  `json:"opponent_destruction,omitempty" db:"opponent_destruction"`
	FetchedAt           time.Time `json:"fetched_at" db:"fetched_at"`
}

type Member struct {
	ID                     string   `json:"member_id" db:"id"`
	WarSnapshotID          string   `json:"war_snapshot_id,omitempty" db:"war_snapshot_id"`
	Side                   string   `json:"side" db:"side"`
	MapPosition            int      `json:"map_position" db:"map_position"`
	PlayerTag              string   `json:"player_tag,omitempty" db:"player_tag"`
	PlayerName             string   `json:"player_name" db:"player_name"`
	THLevel                *int     `json:"th_level,omitempty" db:"th_level"`
	AttacksUsed            *int     `json:"attacks_used,omitempty" db:"attacks_used"`
	BestStarsAgainst       *int     `json:"best_stars_against,omitempty" db:"best_stars_against"`
	BestDestructionAgainst *float64 `json:"best_destruction_against,omitempty" db:"best_destruction_against"`
}

type Target struct {
	ID             string `json:"target_id"`
	WarSnapshotID  string `json:"war_snapshot_id"`
	WarMemberID    string `json:"war_member_id"`
	TargetPosition int    `json:"target_position"`
	TargetName     string `json:"target_name"`
	TargetTH       *int   `json:"target_th,omitempty"`
}

type SaveSnapshotInput struct {
	ID                  string
	ClanTag             string
	OpponentClanTag     string
	WarState            string
	TeamSize            int
	ClanStars           *int
	OpponentStars       *int
	ClanDestruction     *float64
	OpponentDestruction *float64
	FetchedAt           time.Time
	Members             []Member
	Targets             []Target
}

type MemberListResult struct {
	Items []Member
	Total int
}

// CWLGroup represents the Clan War League group from the official CoC API.
type CWLGroup struct {
	State    string     `json:"state"`
	Season   string     `json:"season"`
	ClanTag  string     `json:"clanTag"`
	ClanName string     `json:"clanName"`
	Clans    []CWLClan  `json:"clans"`
	Rounds   []CWLRound `json:"rounds"`
}

// CWLClan is a participating clan in a CWL group.
type CWLClan struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	ClanLevel int    `json:"clanLevel"`
	Members   int    `json:"members"`
	WarWins   int    `json:"warWins,omitempty"`
}

// CWLRound represents one day/round in a CWL group.
type CWLRound struct {
	WarTags []string `json:"warTags"`
	State   string   `json:"state,omitempty"`
}
