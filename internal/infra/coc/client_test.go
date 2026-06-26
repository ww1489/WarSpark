package coc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func TestCurrentWarConvertsCocapiToDomain(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"state":    "inWar",
			"teamSize": 15,
			"clan": map[string]any{
				"tag":                   "#AAA",
				"name":                  "My Clan",
				"stars":                 10,
				"destructionPercentage": 75.5,
				"members": []map[string]any{
					{
						"tag":           "#P1",
						"name":          "Player1",
						"townhallLevel": 14,
						"mapPosition":   1,
						"attacks": []map[string]any{
							{"attackerTag": "#P1", "defenderTag": "#E1", "stars": 3, "destructionPercentage": 100, "order": 1, "duration": 180},
						},
					},
				},
			},
			"opponent": map[string]any{
				"tag":                   "#BBB",
				"name":                  "Enemy",
				"stars":                 5,
				"destructionPercentage": 50.0,
				"members": []map[string]any{
					{"tag": "#E1", "name": "Enemy1", "townhallLevel": 14, "mapPosition": 1},
				},
			},
		})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	war, err := client.CurrentWar(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("CurrentWar error: %v", err)
	}

	if gotPath != "/clans/%23AAA/currentwar" {
		t.Fatalf("path = %q, want /clans/%%23AAA/currentwar", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("auth = %q, want Bearer secret", gotAuth)
	}

	if war.State != "inWar" || war.TeamSize != 15 {
		t.Fatalf("state=%s teamSize=%d", war.State, war.TeamSize)
	}
	if war.Clan.Name != "My Clan" || war.Clan.Stars != 10 {
		t.Fatalf("clan: %+v", war.Clan)
	}
	if war.Clan.DestructionPercentage != 75.5 {
		t.Fatalf("destruction = %v, want 75.5", war.Clan.DestructionPercentage)
	}
	if len(war.Clan.Members) != 1 {
		t.Fatalf("members = %d, want 1", len(war.Clan.Members))
	}
	m := war.Clan.Members[0]
	if m.TownHallLevel != 14 || m.MapPosition != 1 {
		t.Fatalf("member: %+v", m)
	}
	if len(m.Attacks) != 1 || m.Attacks[0].Stars != 3 || m.Attacks[0].DestructionPercentage != 100 {
		t.Fatalf("attacks: %+v", m.Attacks)
	}
	if war.Opponent.Tag != "#BBB" || war.Opponent.DestructionPercentage != 50.0 {
		t.Fatalf("opponent: %+v", war.Opponent)
	}
}

func TestCWLGroupConvertsMembersCountAndRounds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"state":  "preparation",
			"season": "2026-06",
			"clans": []map[string]any{
				{"tag": "#C1", "name": "Clan1", "clanLevel": 10, "members": []map[string]any{{"tag": "#M1"}, {"tag": "#M2"}}},
				{"tag": "#C2", "name": "Clan2", "clanLevel": 8, "members": []map[string]any{{"tag": "#M3"}}},
			},
			"rounds": []map[string]any{
				{"warTags": []string{"#W1", "#W2"}},
				{"warTags": []string{"#W3", "#W4"}},
			},
		})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	group, err := client.CWLGroup(context.Background(), "#C1")
	if err != nil {
		t.Fatalf("CWLGroup error: %v", err)
	}

	if group.State != "preparation" || group.Season != "2026-06" {
		t.Fatalf("state=%s season=%s", group.State, group.Season)
	}
	if group.ClanTag != "" || group.ClanName != "" {
		t.Fatalf("adapter should leave ClanTag/ClanName empty, got %q/%q", group.ClanTag, group.ClanName)
	}
	if len(group.Clans) != 2 {
		t.Fatalf("clans = %d, want 2", len(group.Clans))
	}
	if group.Clans[0].Members != 2 || group.Clans[1].Members != 1 {
		t.Fatalf("members count: %d, %d (want 2, 1)", group.Clans[0].Members, group.Clans[1].Members)
	}
	if len(group.Rounds) != 2 || len(group.Rounds[0].WarTags) != 2 {
		t.Fatalf("rounds: %+v", group.Rounds)
	}
}

func TestCurrentWarRequiresToken(t *testing.T) {
	client := New(cocapi.Config{BaseURL: "https://example.com", Timeout: time.Second})
	_, err := client.CurrentWar(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeAPINotConfigured {
		t.Fatalf("error code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeAPINotConfigured)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		name    string
		cocapi  error
		wantCod string
	}{
		{"not configured", cocapi.ErrAPINotConfigured, dmerrors.ErrCodeAPINotConfigured},
		{"access denied", cocapi.ErrAPIAccessDenied, dmerrors.ErrCodeAPIAccessDenied},
		{"not found", cocapi.ErrNotFound, dmerrors.ErrCodeWarNotFound},
		{"invalid tag", cocapi.ErrInvalidTag, dmerrors.ErrCodeInvalidTag},
		{"response invalid", cocapi.ErrAPIResponseInvalid, dmerrors.ErrCodeAPIResponseInvalid},
		{"request failed", cocapi.ErrAPIRequestFailed, dmerrors.ErrCodeAPIRequestFailed},
		{"rate limited", cocapi.ErrRateLimited, dmerrors.ErrCodeAPIRequestFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := mapError(tc.cocapi)
			if dmerrors.Code(mapped) != tc.wantCod {
				t.Fatalf("code = %q, want %q", dmerrors.Code(mapped), tc.wantCod)
			}
			if !errors.Is(mapped, tc.cocapi) {
				t.Fatalf("mapped error should wrap original: %v", mapped)
			}
		})
	}
}

func TestCurrentWarMapsNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"reason": "notFound", "message": "clan not found"})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	_, err := client.CurrentWar(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeWarNotFound {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeWarNotFound)
	}
}

func TestCurrentWarMapsAccessDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{"reason": "accessDenied", "message": "denied"})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	_, err := client.CurrentWar(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeAPIAccessDenied {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeAPIAccessDenied)
	}
}
