package league

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type League struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LeagueListResponse struct {
	Items  []League `json:"items"`
	Paging Paging   `json:"paging"`
}

type LeagueSeason struct {
	ID string `json:"id"`
}

type LeagueSeasonListResponse struct {
	Items  []LeagueSeason `json:"items"`
	Paging Paging         `json:"paging"`
}

type LeagueSeasonRankingEntry struct {
	Tag          string   `json:"tag"`
	Name         string   `json:"name"`
	ExpLevel     int      `json:"expLevel"`
	Trophies     int      `json:"trophies"`
	Rank         int      `json:"rank"`
	PreviousRank int      `json:"previousRank,omitempty"`
	AttackWins   int      `json:"attackWins,omitempty"`
	DefenseWins  int      `json:"defenseWins,omitempty"`
	Clan         *ClanRef `json:"clan,omitempty"`
}

type ClanRef struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type LeagueSeasonRankingListResponse struct {
	Items  []LeagueSeasonRankingEntry `json:"items"`
	Paging Paging                     `json:"paging"`
}

type LeagueTier struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LeagueTierListResponse struct {
	Items  []LeagueTier `json:"items"`
	Paging Paging       `json:"paging"`
}

type WarLeague struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WarLeagueListResponse struct {
	Items  []WarLeague `json:"items"`
	Paging Paging      `json:"paging"`
}

type LeagueSeasonResult struct {
	LeagueSeasonID int64 `json:"leagueSeasonId"`
	LeagueTierID   int   `json:"leagueTierId,omitempty"`
	Trophies       int   `json:"trophies"`
	Placement      int   `json:"placement,omitempty"`
	AttackWins     int   `json:"attackWins,omitempty"`
	AttackLosses   int   `json:"attackLosses,omitempty"`
	AttackStars    int   `json:"attackStars,omitempty"`
	DefenseWins    int   `json:"defenseWins,omitempty"`
	DefenseLosses  int   `json:"defenseLosses,omitempty"`
	DefenseStars   int   `json:"defenseStars,omitempty"`
	MaxBattles     int   `json:"maxBattles,omitempty"`
}

type LeagueSeasonResultListResponse struct {
	Items  []LeagueSeasonResult `json:"items"`
	Paging Paging               `json:"paging"`
}
