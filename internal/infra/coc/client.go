package coc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	appconfig "github.com/ww1489/WarSpark/internal/config"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

func New(cfg appconfig.CoCConfig) *Client {
	return &Client{
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		apiToken: cfg.APIToken,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *Client) CurrentWar(ctx context.Context, clanTag string) (wardomain.CurrentWar, error) {
	if strings.TrimSpace(c.apiToken) == "" {
		return wardomain.CurrentWar{}, wardomain.NewError(wardomain.ErrorAPINotConfigured, "clash of clans api token is not configured")
	}

	endpoint := c.baseURL + "/clans/" + url.PathEscape(clanTag) + "/currentwar"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return wardomain.CurrentWar{}, wardomain.WrapError(wardomain.ErrorAPIRequestFailed, "build clash of clans api request", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiToken)
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return wardomain.CurrentWar{}, wardomain.WrapError(wardomain.ErrorAPIRequestFailed, "clash of clans api request failed", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusUnauthorized {
		return wardomain.CurrentWar{}, wardomain.NewError(wardomain.ErrorAPIAccessDenied, "clash of clans api access denied")
	}
	if response.StatusCode == http.StatusNotFound {
		return wardomain.CurrentWar{}, wardomain.NewError(wardomain.ErrorWarNotFound, "current war not found")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return wardomain.CurrentWar{}, wardomain.NewError(wardomain.ErrorAPIRequestFailed, fmt.Sprintf("clash of clans api returned status %d", response.StatusCode))
	}

	var currentWar wardomain.CurrentWar
	if err := json.NewDecoder(response.Body).Decode(&currentWar); err != nil {
		return wardomain.CurrentWar{}, wardomain.WrapError(wardomain.ErrorAPIResponseInvalid, "decode clash of clans api response", err)
	}
	return currentWar, nil
}


// CWLGroup fetches the current Clan War League group for a clan.
func (c *Client) CWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error) {
	if strings.TrimSpace(c.apiToken) == "" {
		return wardomain.CWLGroup{}, wardomain.NewError(wardomain.ErrorAPINotConfigured, "clash of clans api token is not configured")
	}

	endpoint := c.baseURL + "/clans/" + url.PathEscape(clanTag) + "/currentwar/leaguegroup"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return wardomain.CWLGroup{}, wardomain.WrapError(wardomain.ErrorAPIRequestFailed, "build cwl api request", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiToken)
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return wardomain.CWLGroup{}, wardomain.WrapError(wardomain.ErrorAPIRequestFailed, "cwl api request failed", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusUnauthorized {
		return wardomain.CWLGroup{}, wardomain.NewError(wardomain.ErrorAPIAccessDenied, "clash of clans api access denied")
	}
	if response.StatusCode == http.StatusNotFound {
		return wardomain.CWLGroup{}, wardomain.NewError(wardomain.ErrorWarNotFound, "cwl group not found")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return wardomain.CWLGroup{}, wardomain.NewError(wardomain.ErrorAPIRequestFailed, fmt.Sprintf("cwl api returned status %d", response.StatusCode))
	}

	var raw struct {
		State  string `json:"state"`
		Season string `json:"season"`
		Clans  []struct {
			Tag       string `json:"tag"`
			Name      string `json:"name"`
			ClanLevel int    `json:"clanLevel"`
			Members   []struct {
				Tag           string `json:"tag"`
				Name          string `json:"name"`
				TownHallLevel int    `json:"townHallLevel"`
			} `json:"members"`
			WarWins int `json:"warWins"`
		} `json:"clans"`
		Rounds []struct {
			WarTags []string `json:"warTags"`
		} `json:"rounds"`
	}
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return wardomain.CWLGroup{}, wardomain.WrapError(wardomain.ErrorAPIResponseInvalid, "decode cwl api response", err)
	}

	group := wardomain.CWLGroup{
		State:  raw.State,
		Season: raw.Season,
	}
	for _, cl := range raw.Clans {
		group.Clans = append(group.Clans, wardomain.CWLClan{
			Tag:       cl.Tag,
			Name:      cl.Name,
			ClanLevel: cl.ClanLevel,
			Members:   len(cl.Members),
			WarWins:   cl.WarWins,
		})
	}
	for _, r := range raw.Rounds {
		group.Rounds = append(group.Rounds, wardomain.CWLRound{
			WarTags: r.WarTags,
		})
	}
	return group, nil
}

func TimeoutOrDefault(value time.Duration) time.Duration {
	if value <= 0 {
		return 10 * time.Second
	}
	return value
}
