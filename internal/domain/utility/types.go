package utility

type GoldPassSeason struct {
	EndTime   string `json:"endTime"`
	StartTime string `json:"startTime"`
}

type LocationDetail struct {
	CountryCode   string `json:"countryCode,omitempty"`
	ID            int    `json:"id"`
	IsCountry     bool   `json:"isCountry"`
	LocalizedName string `json:"localizedName"`
	Name          string `json:"name"`
}

type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type VerifyTokenResponse struct {
	Status string `json:"status"`
	Tag    string `json:"tag"`
	Token  string `json:"token"`
}

type PlayerLeagueGroup struct {
	AttackLogs  []LeagueBattleLogEntry `json:"attackLogs"`
	DefenseLogs []LeagueBattleLogEntry `json:"defenseLogs"`
	Members     []LeagueGroupMember    `json:"members"`
}

type LeagueBattleLogEntry struct {
	CreationTime          string `json:"creationTime"`
	DestructionPercentage int    `json:"destructionPercentage"`
	OpponentName          string `json:"opponentName"`
	OpponentPlayerTag     string `json:"opponentPlayerTag"`
	Stars                 int    `json:"stars"`
	Trophies              int    `json:"trophies"`
}

type LeagueGroupMember struct {
	AttackLoseCount  int    `json:"attackLoseCount"`
	AttackWinCount   int    `json:"attackWinCount"`
	DefenseLoseCount int    `json:"defenseLoseCount"`
	DefenseWinCount  int    `json:"defenseWinCount"`
	LeagueTrophies   int    `json:"leagueTrophies"`
	ClanName         string `json:"clanName"`
	ClanTag          string `json:"clanTag"`
	PlayerName       string `json:"playerName"`
	PlayerTag        string `json:"playerTag"`
}

type ClanSearchParams struct {
	Name          string `form:"name"`
	WarFrequency  string `form:"warFrequency"`
	LocationID    int    `form:"locationId"`
	MinMembers    int    `form:"minMembers"`
	MaxMembers    int    `form:"maxMembers"`
	MinClanPoints int    `form:"minClanPoints"`
	MinClanLevel  int    `form:"minClanLevel"`
	Limit         int    `form:"limit"`
	After         string `form:"after"`
	Before        string `form:"before"`
	LabelIds      string `form:"labelIds"`
}
