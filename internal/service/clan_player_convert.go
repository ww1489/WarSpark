package service

import (
	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func toDomainLabels(labels cocapi.LabelList) []clandomain.Label {
	result := make([]clandomain.Label, 0, len(labels))
	for _, l := range labels {
		result = append(result, clandomain.Label{
			ID:   l.ID,
			Name: string(l.Name),
		})
	}
	return result
}

func toDomainLocation(loc cocapi.Location) clandomain.Location {
	return clandomain.Location{
		ID:          loc.ID,
		Name:        string(loc.Name),
		CountryCode: loc.CountryCode,
	}
}

func toDomainLeagueRef(league cocapi.League) clandomain.LeagueRef {
	return clandomain.LeagueRef{
		ID:   league.ID,
		Name: string(league.Name),
	}
}

func toDomainHeroLevels(heroes cocapi.PlayerItemLevelList) []clandomain.HeroLevel {
	result := make([]clandomain.HeroLevel, 0, len(heroes))
	for _, h := range heroes {
		result = append(result, clandomain.HeroLevel{
			Name:     string(h.Name),
			Level:    h.Level,
			MaxLevel: h.MaxLevel,
			Village:  h.Village,
		})
	}
	return result
}

func toDomainAchievements(achievements cocapi.PlayerAchievementProgressList) []clandomain.AchievementProgress {
	result := make([]clandomain.AchievementProgress, 0, len(achievements))
	for _, a := range achievements {
		result = append(result, clandomain.AchievementProgress{
			Name:    string(a.Name),
			Stars:   a.Stars,
			Target:  a.Target,
			Value:   a.Value,
			Village: a.Village,
			Info:    string(a.Info),
		})
	}
	return result
}
