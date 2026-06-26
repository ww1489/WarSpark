package capital

import "encoding/json"

type CapitalRaidSeason struct {
	AttackLog               json.RawMessage `json:"attackLog"`
	CapitalTotalLoot        int             `json:"capitalTotalLoot"`
	DefenseLog              json.RawMessage `json:"defenseLog"`
	DefensiveReward         int             `json:"defensiveReward"`
	EndTime                 string          `json:"endTime"`
	EnemyDistrictsDestroyed int             `json:"enemyDistrictsDestroyed"`
	Members                 json.RawMessage `json:"members"`
	OffensiveReward         int             `json:"offensiveReward"`
	RaidsCompleted          int             `json:"raidsCompleted"`
	StartTime               string          `json:"startTime"`
	State                   string          `json:"state"`
	TotalAttacks            int             `json:"totalAttacks"`
}

type CapitalRaidSeasonListResponse struct {
	Items  []CapitalRaidSeason `json:"items"`
	Paging Paging              `json:"paging"`
}

type CapitalLeague struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CapitalLeagueListResponse struct {
	Items  []CapitalLeague `json:"items"`
	Paging Paging          `json:"paging"`
}

type BuilderBaseLeague struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BuilderBaseLeagueListResponse struct {
	Items  []BuilderBaseLeague `json:"items"`
	Paging Paging              `json:"paging"`
}

type Paging struct {
	Cursors PagingCursors `json:"cursors"`
}

type PagingCursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}
