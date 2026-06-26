package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

var clanTagPattern = regexp.MustCompile(`^#?[A-Z0-9]{3,16}$`)

type WarAPIClient interface {
	CurrentWar(ctx context.Context, clanTag string) (wardomain.CurrentWar, error)
	CWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error)
}

type WarRepository interface {
	SaveSnapshot(ctx context.Context, input wardomain.SaveSnapshotInput) (wardomain.Snapshot, error)
	ListMembers(ctx context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error)
}

type WarCache interface {
	GetCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, bool, error)
	SetCurrentWar(ctx context.Context, clanTag string, snapshot wardomain.Snapshot, ttl time.Duration) error
	GetCWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, bool, error)
	SetCWLGroup(ctx context.Context, clanTag string, group wardomain.CWLGroup, ttl time.Duration) error
	GetWarLog(ctx context.Context, clanTag string) (cocapi.ClanWarLogResponse, bool, error)
	SetWarLog(ctx context.Context, clanTag string, resp cocapi.ClanWarLogResponse, ttl time.Duration) error
	GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, bool, error)
	SetCWLWar(ctx context.Context, warTag string, resp cocapi.ClanWar, ttl time.Duration) error
}

type WarServiceOptions struct {
	Cache              WarCache
	CurrentWarCacheTTL time.Duration
	CWLGroupCacheTTL   time.Duration
	CocapiClient       *cocapi.Client
}

type WarService struct {
	client             WarAPIClient
	repository         WarRepository
	cache              WarCache
	currentWarCacheTTL time.Duration
	cwlGroupCacheTTL   time.Duration
	capi               *cocapi.Client
	now                func() time.Time
}

func NewWarService(client WarAPIClient, repository WarRepository, options ...WarServiceOptions) *WarService {
	var option WarServiceOptions
	if len(options) > 0 {
		option = options[0]
	}
	if option.CurrentWarCacheTTL <= 0 {
		option.CurrentWarCacheTTL = 2 * time.Minute
	}
	if option.CWLGroupCacheTTL <= 0 {
		option.CWLGroupCacheTTL = 5 * time.Minute
	}
	return &WarService{
		client:             client,
		repository:         repository,
		cache:              option.Cache,
		currentWarCacheTTL: option.CurrentWarCacheTTL,
		cwlGroupCacheTTL:   option.CWLGroupCacheTTL,
		capi:               option.CocapiClient,
		now:                time.Now,
	}
}

func (s *WarService) FetchCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return wardomain.Snapshot{}, err
	}

	if s.cache != nil {
		snapshot, ok, err := s.cache.GetCurrentWar(ctx, normalizedTag)
		if err == nil && ok {
			return snapshot, nil
		}
	}

	currentWar, err := s.client.CurrentWar(ctx, normalizedTag)
	if err != nil {
		return wardomain.Snapshot{}, err
	}

	input := buildSnapshotInput(currentWar, normalizedTag, s.now().UTC())
	snapshot, err := s.repository.SaveSnapshot(ctx, input)
	if err != nil {
		return wardomain.Snapshot{}, err
	}
	if s.cache != nil {
		_ = s.cache.SetCurrentWar(ctx, normalizedTag, snapshot, s.currentWarCacheTTL)
	}
	return snapshot, nil
}

func (s *WarService) FetchCWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return wardomain.CWLGroup{}, err
	}

	if s.cache != nil {
		group, ok, err := s.cache.GetCWLGroup(ctx, normalizedTag)
		if err == nil && ok {
			return group, nil
		}
	}

	group, err := s.client.CWLGroup(ctx, normalizedTag)
	if err != nil {
		return wardomain.CWLGroup{}, err
	}
	if group.ClanTag == "" {
		group.ClanTag = normalizedTag
	}
	for _, clan := range group.Clans {
		if clan.Tag == normalizedTag {
			group.ClanName = clan.Name
			break
		}
	}
	if s.cache != nil {
		_ = s.cache.SetCWLGroup(ctx, normalizedTag, group, s.cwlGroupCacheTTL)
	}
	return group, nil
}

func (s *WarService) ListMembers(ctx context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error) {
	return s.repository.ListMembers(ctx, snapshotID, side, pagination)
}

func (s *WarService) GetWarLog(ctx context.Context, clanTag string, limit int, after, before string) (cocapi.ClanWarLogResponse, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return cocapi.ClanWarLogResponse{}, err
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetWarLog(ctx, normalizedTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	resp, err := s.capi.GetClanWarLog(ctx, normalizedTag, cocapi.QueryGetClanWarLog{
		Limit:  limit,
		After:  after,
		Before: before,
	})
	if err != nil {
		return cocapi.ClanWarLogResponse{}, mapCocapiErrorToWarError(err)
	}
	if s.cache != nil {
		_ = s.cache.SetWarLog(ctx, normalizedTag, resp, s.currentWarCacheTTL)
	}
	return resp, nil
}

func (s *WarService) GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, error) {
	normalizedTag, err := NormalizeClanTag(warTag)
	if err != nil {
		return cocapi.ClanWar{}, err
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetCWLWar(ctx, normalizedTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	resp, err := s.capi.GetClanWarLeagueWar(ctx, normalizedTag)
	if err != nil {
		return cocapi.ClanWar{}, mapCocapiErrorToWarError(err)
	}
	if s.cache != nil {
		_ = s.cache.SetCWLWar(ctx, normalizedTag, resp, s.cwlGroupCacheTTL)
	}
	return resp, nil
}

func mapCocapiErrorToWarError(err error) error {
	if errors.Is(err, cocapi.ErrNotFound) {
		return wardomain.NewError(wardomain.ErrorWarNotFound, err.Error())
	}
	if errors.Is(err, cocapi.ErrInvalidTag) {
		return wardomain.NewError(wardomain.ErrorInvalidTag, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPINotConfigured) {
		return wardomain.NewError(wardomain.ErrorAPINotConfigured, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPIAccessDenied) {
		return wardomain.NewError(wardomain.ErrorAPIAccessDenied, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPIResponseInvalid) {
		return wardomain.NewError(wardomain.ErrorAPIResponseInvalid, err.Error())
	}
	return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
}

func NormalizeClanTag(value string) (string, error) {
	tag := strings.ToUpper(strings.TrimSpace(value))
	if tag == "" {
		return "", wardomain.NewError(wardomain.ErrorInvalidTag, "clan tag is required")
	}
	if !strings.HasPrefix(tag, "#") {
		tag = "#" + tag
	}
	if !clanTagPattern.MatchString(tag) {
		return "", wardomain.NewError(wardomain.ErrorInvalidTag, "invalid clan tag")
	}
	return tag, nil
}

func buildSnapshotInput(currentWar wardomain.CurrentWar, requestedClanTag string, fetchedAt time.Time) wardomain.SaveSnapshotInput {
	clanStars := currentWar.Clan.Stars
	opponentStars := currentWar.Opponent.Stars
	clanDestruction := currentWar.Clan.DestructionPercentage
	opponentDestruction := currentWar.Opponent.DestructionPercentage

	members := make([]wardomain.Member, 0, len(currentWar.Clan.Members)+len(currentWar.Opponent.Members))
	targets := make([]wardomain.Target, 0, len(currentWar.Opponent.Members))
	attackIndex := attackIndex(currentWar)

	for _, member := range currentWar.Clan.Members {
		members = append(members, warMember("clan", member, attackIndex[normalizeAPIPlayerTag(member.Tag)]))
	}
	for _, member := range currentWar.Opponent.Members {
		converted := warMember("opponent", member, attackIndex[normalizeAPIPlayerTag(member.Tag)])
		members = append(members, converted)
		targets = append(targets, wardomain.Target{
			ID:             uuid.NewString(),
			WarMemberID:    converted.ID,
			TargetPosition: converted.MapPosition,
			TargetName:     converted.PlayerName,
			TargetTH:       converted.THLevel,
		})
	}

	return wardomain.SaveSnapshotInput{
		ID:                  uuid.NewString(),
		ClanTag:             requestedClanTag,
		OpponentClanTag:     currentWar.Opponent.Tag,
		WarState:            currentWar.State,
		TeamSize:            currentWar.TeamSize,
		ClanStars:           &clanStars,
		OpponentStars:       &opponentStars,
		ClanDestruction:     &clanDestruction,
		OpponentDestruction: &opponentDestruction,
		FetchedAt:           fetchedAt,
		Members:             members,
		Targets:             targets,
	}
}

func warMember(side string, member wardomain.WarMember, incoming []wardomain.WarAttack) wardomain.Member {
	thLevel := member.TownHallLevel
	attacksUsed := len(member.Attacks)
	result := wardomain.Member{
		ID:          uuid.NewString(),
		Side:        side,
		MapPosition: member.MapPosition,
		PlayerTag:   normalizeAPIPlayerTag(member.Tag),
		PlayerName:  member.Name,
		THLevel:     &thLevel,
		AttacksUsed: &attacksUsed,
	}
	if len(incoming) > 0 {
		bestStars, bestDestruction := bestDefenseResult(incoming)
		result.BestStarsAgainst = &bestStars
		result.BestDestructionAgainst = &bestDestruction
	}
	return result
}

func attackIndex(currentWar wardomain.CurrentWar) map[string][]wardomain.WarAttack {
	index := make(map[string][]wardomain.WarAttack)
	for _, member := range currentWar.Clan.Members {
		for _, attack := range member.Attacks {
			defenderTag := normalizeAPIPlayerTag(attack.DefenderTag)
			index[defenderTag] = append(index[defenderTag], attack)
		}
	}
	for _, member := range currentWar.Opponent.Members {
		for _, attack := range member.Attacks {
			defenderTag := normalizeAPIPlayerTag(attack.DefenderTag)
			index[defenderTag] = append(index[defenderTag], attack)
		}
	}
	return index
}

func bestDefenseResult(attacks []wardomain.WarAttack) (int, float64) {
	bestStars := attacks[0].Stars
	bestDestruction := attacks[0].DestructionPercentage
	for _, attack := range attacks[1:] {
		if attack.Stars > bestStars || (attack.Stars == bestStars && attack.DestructionPercentage > bestDestruction) {
			bestStars = attack.Stars
			bestDestruction = attack.DestructionPercentage
		}
	}
	return bestStars, bestDestruction
}

func normalizeAPIPlayerTag(value string) string {
	tag := strings.ToUpper(strings.TrimSpace(value))
	if tag != "" && !strings.HasPrefix(tag, "#") {
		tag = "#" + tag
	}
	return tag
}
