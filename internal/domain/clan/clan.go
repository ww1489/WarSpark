package clan

type ClanDetail struct {
	Clan    ClanOverview        `json:"clan"`
	Members []ClanMemberSummary `json:"members"`
}

type ClanOverview struct {
	Tag            string     `json:"tag"`
	Name           string     `json:"name"`
	ClanLevel      int        `json:"clanLevel"`
	Description    string     `json:"description,omitempty"`
	Members        int        `json:"members"`
	ClanPoints     int        `json:"clanPoints,omitempty"`
	WarWins        int        `json:"warWins,omitempty"`
	WarLosses      int        `json:"warLosses,omitempty"`
	WarTies        int        `json:"warTies,omitempty"`
	WarWinStreak   int        `json:"warWinStreak,omitempty"`
	WarFrequency   string     `json:"warFrequency,omitempty"`
	Type           string     `json:"type,omitempty"`
	IsWarLogPublic bool       `json:"isWarLogPublic,omitempty"`
	BadgeURLs      any        `json:"badgeUrls,omitempty"`
	Labels         []Label    `json:"labels,omitempty"`
	Location       *Location  `json:"location,omitempty"`
	WarLeague      *LeagueRef `json:"warLeague,omitempty"`
}

type ClanMemberSummary struct {
	Tag               string     `json:"tag"`
	Name              string     `json:"name"`
	TownHallLevel     int        `json:"townHallLevel"`
	ExpLevel          int        `json:"expLevel"`
	Role              string     `json:"role"`
	Trophies          int        `json:"trophies"`
	ClanRank          int        `json:"clanRank"`
	Donations         int        `json:"donations"`
	DonationsReceived int        `json:"donationsReceived"`
	League            *LeagueRef `json:"league,omitempty"`
}

type PlayerOverview struct {
	Tag                 string                `json:"tag"`
	Name                string                `json:"name"`
	TownHallLevel       int                   `json:"townHallLevel"`
	TownHallWeaponLevel int                   `json:"townHallWeaponLevel,omitempty"`
	ExpLevel            int                   `json:"expLevel"`
	Role                string                `json:"role,omitempty"`
	WarStars            int                   `json:"warStars,omitempty"`
	AttackWins          int                   `json:"attackWins,omitempty"`
	DefenseWins         int                   `json:"defenseWins,omitempty"`
	Trophies            int                   `json:"trophies"`
	BestTrophies        int                   `json:"bestTrophies,omitempty"`
	WarPreference       string                `json:"warPreference,omitempty"`
	Clan                *PlayerClanInfo       `json:"clan,omitempty"`
	Heroes              []HeroLevel           `json:"heroes,omitempty"`
	Achievements        []AchievementProgress `json:"achievements,omitempty"`
	Labels              []Label               `json:"labels,omitempty"`
	League              *LeagueRef            `json:"league,omitempty"`
}

type PlayerClanInfo struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	ClanLevel int    `json:"clanLevel"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type HeroLevel struct {
	Name     string `json:"name"`
	Level    int    `json:"level"`
	MaxLevel int    `json:"maxLevel,omitempty"`
	Village  string `json:"village,omitempty"`
}

type AchievementProgress struct {
	Name    string `json:"name"`
	Stars   int    `json:"stars"`
	Target  int    `json:"target"`
	Value   int    `json:"value"`
	Village string `json:"village,omitempty"`
	Info    string `json:"info,omitempty"`
}

type BattleLogSummary struct {
	Items  []BattleLogEntry `json:"items"`
	Paging Paging           `json:"paging"`
}

type BattleLogEntry struct {
	ArmyShareCode         string `json:"armyShareCode,omitempty"`
	Attack                bool   `json:"attack,omitempty"`
	BattleTime            int    `json:"battleTime,omitempty"`
	BattleTimestamp       string `json:"battleTimestamp,omitempty"`
	BattleType            string `json:"battleType,omitempty"`
	DestructionPercentage int    `json:"destructionPercentage,omitempty"`
	OpponentName          string `json:"opponentName,omitempty"`
	OpponentPlayerTag     string `json:"opponentPlayerTag,omitempty"`
	OpponentTownHallLevel int    `json:"opponentTownHallLevel,omitempty"`
	Stars                 int    `json:"stars,omitempty"`
}

type Label struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Location struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CountryCode string `json:"countryCode,omitempty"`
}

type LeagueRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Paging struct {
	Cursors struct {
		After  string `json:"after,omitempty"`
		Before string `json:"before,omitempty"`
	} `json:"cursors,omitempty"`
}
