package ranking

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type Location struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	CountryCode   string `json:"countryCode,omitempty"`
	IsCountry     bool   `json:"isCountry,omitempty"`
	LocalizedName string `json:"localizedName,omitempty"`
}

type LocationListResponse struct {
	Items  []Location `json:"items"`
	Paging Paging     `json:"paging"`
}

type ClanRankingEntry struct {
	Tag          string    `json:"tag"`
	Name         string    `json:"name"`
	ClanLevel    int       `json:"clanLevel"`
	ClanPoints   int       `json:"clanPoints"`
	Members      int       `json:"members"`
	Rank         int       `json:"rank"`
	PreviousRank int       `json:"previousRank,omitempty"`
	BadgeURLs    any       `json:"badgeUrls,omitempty"`
	Location     *Location `json:"location,omitempty"`
}

type ClanRankingListResponse struct {
	Items  []ClanRankingEntry `json:"items"`
	Paging Paging             `json:"paging"`
}

type PlayerRankingEntry struct {
	Tag          string   `json:"tag"`
	Name         string   `json:"name"`
	ExpLevel     int      `json:"expLevel"`
	Trophies     int      `json:"trophies"`
	Rank         int      `json:"rank"`
	PreviousRank int      `json:"previousRank,omitempty"`
	AttackWins   int      `json:"attackWins,omitempty"`
	DefenseWins  int      `json:"defenseWins,omitempty"`
	Clan         *ClanRef `json:"clan,omitempty"`
	LeagueTier   any      `json:"leagueTier,omitempty"`
}

type ClanRef struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type PlayerRankingListResponse struct {
	Items  []PlayerRankingEntry `json:"items"`
	Paging Paging               `json:"paging"`
}

type ClanCapitalRankingEntry struct {
	Tag               string    `json:"tag"`
	Name              string    `json:"name"`
	ClanLevel         int       `json:"clanLevel"`
	ClanCapitalPoints int       `json:"clanCapitalPoints"`
	Rank              int       `json:"rank"`
	PreviousRank      int       `json:"previousRank,omitempty"`
	Members           int       `json:"members"`
	BadgeURLs         any       `json:"badgeUrls,omitempty"`
	Location          *Location `json:"location,omitempty"`
}

type ClanCapitalRankingListResponse struct {
	Items  []ClanCapitalRankingEntry `json:"items"`
	Paging Paging                    `json:"paging"`
}

type ClanBuilderBaseRankingEntry struct {
	Tag                   string    `json:"tag"`
	Name                  string    `json:"name"`
	ClanLevel             int       `json:"clanLevel"`
	ClanBuilderBasePoints int       `json:"clanBuilderBasePoints"`
	Rank                  int       `json:"rank"`
	PreviousRank          int       `json:"previousRank,omitempty"`
	Members               int       `json:"members"`
	BadgeURLs             any       `json:"badgeUrls,omitempty"`
	Location              *Location `json:"location,omitempty"`
}

type ClanBuilderBaseRankingListResponse struct {
	Items  []ClanBuilderBaseRankingEntry `json:"items"`
	Paging Paging                        `json:"paging"`
}

type PlayerBuilderBaseRankingEntry struct {
	Tag                 string   `json:"tag"`
	Name                string   `json:"name"`
	ExpLevel            int      `json:"expLevel"`
	BuilderBaseTrophies int      `json:"builderBaseTrophies"`
	Rank                int      `json:"rank"`
	PreviousRank        int      `json:"previousRank,omitempty"`
	Clan                *ClanRef `json:"clan,omitempty"`
}

type PlayerBuilderBaseRankingListResponse struct {
	Items  []PlayerBuilderBaseRankingEntry `json:"items"`
	Paging Paging                          `json:"paging"`
}
