package service

import (
	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func toDomainClanOverview(c cocapi.Clan) clandomain.ClanOverview {
	overview := clandomain.ClanOverview{
		Tag:            c.Tag,
		Name:           c.Name,
		ClanLevel:      c.ClanLevel,
		Description:    c.Description,
		Members:        c.Members,
		ClanPoints:     c.ClanPoints,
		WarWins:        c.WarWins,
		WarLosses:      c.WarLosses,
		WarTies:        c.WarTies,
		WarWinStreak:   c.WarWinStreak,
		WarFrequency:   c.WarFrequency,
		Type:           c.Type,
		IsWarLogPublic: c.IsWarLogPublic,
		BadgeURLs:      c.BadgeURLs,
		Labels:         toDomainLabels(c.Labels),
	}
	if c.Location.ID != 0 {
		loc := toDomainLocation(c.Location)
		overview.Location = &loc
	}
	if c.WarLeague.ID != 0 {
		league := toDomainLeagueRef(cocapi.League{ID: c.WarLeague.ID, Name: c.WarLeague.Name})
		overview.WarLeague = &league
	}
	return overview
}

func toDomainClanMembers(members []cocapi.ClanMember) []clandomain.ClanMemberSummary {
	result := make([]clandomain.ClanMemberSummary, 0, len(members))
	for _, m := range members {
		summary := clandomain.ClanMemberSummary{
			Tag:               m.Tag,
			Name:              m.Name,
			TownHallLevel:     m.TownHallLevel,
			ExpLevel:          m.ExpLevel,
			Role:              m.Role,
			Trophies:          m.Trophies,
			ClanRank:          m.ClanRank,
			Donations:         m.Donations,
			DonationsReceived: m.DonationsReceived,
		}
		if m.League.ID != 0 {
			league := toDomainLeagueRef(cocapi.League(m.League))
			summary.League = &league
		}
		result = append(result, summary)
	}
	return result
}
