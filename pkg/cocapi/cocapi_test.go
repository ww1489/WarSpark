package cocapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDefaultsBaseURLAndTimeout(t *testing.T) {
	c := New(Config{APIToken: "t"})
	if c.baseURL != "https://api.clashofclans.com/v1" {
		t.Fatalf("default base url = %q", c.baseURL)
	}
	if c.httpClient.Timeout != 10*time.Second {
		t.Fatalf("default timeout = %v", c.httpClient.Timeout)
	}
}

func TestNewTrimsTrailingSlash(t *testing.T) {
	c := New(Config{BaseURL: "https://api.clashofclans.com/v1/", APIToken: "t"})
	if c.baseURL != "https://api.clashofclans.com/v1" {
		t.Fatalf("base url not trimmed: %q", c.baseURL)
	}
}

func TestDoRequiresToken(t *testing.T) {
	c := New(Config{BaseURL: "https://example.com"})
	var result Clan
	err := c.Get(context.Background(), "/clans/{}", &result, "#ABC")
	if !errors.Is(err, ErrAPINotConfigured) {
		t.Fatalf("expected ErrAPINotConfigured, got %v", err)
	}
}

func TestGetClanSendsBearerAndEscapedTag(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(Clan{Tag: "#2PP", Name: "Test"})
	}))
	defer server.Close()

	c := New(Config{BaseURL: server.URL, APIToken: "secret"})
	clan, err := c.GetClan(context.Background(), "#2PP")
	if err != nil {
		t.Fatalf("GetClan error: %v", err)
	}
	if gotPath != "/clans/%232PP" {
		t.Fatalf("expected escaped path /clans/%%232PP, got %q", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("expected bearer token, got %q", gotAuth)
	}
	if clan.Name != "Test" {
		t.Fatalf("unexpected clan: %+v", clan)
	}
}

func TestGetClanNormalizesTagWithoutHash(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_ = json.NewEncoder(w).Encode(Clan{})
	}))
	defer server.Close()

	c := New(Config{BaseURL: server.URL, APIToken: "secret"})
	if _, err := c.GetClan(context.Background(), "2pp"); err != nil {
		t.Fatalf("GetClan error: %v", err)
	}
	if gotPath != "/clans/%232PP" {
		t.Fatalf("expected normalized+escaped tag, got %q", gotPath)
	}
}

func TestGetLeagueConvertsIntPathParam(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_ = json.NewEncoder(w).Encode(League{ID: 29000022})
	}))
	defer server.Close()

	c := New(Config{BaseURL: server.URL, APIToken: "secret"})
	if _, err := c.GetLeague(context.Background(), "29000022"); err != nil {
		t.Fatalf("GetLeague error: %v", err)
	}
	if gotPath != "/leagues/29000022" {
		t.Fatalf("expected /leagues/29000022, got %q", gotPath)
	}
}

func TestSearchClansSendsQueryParams(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(ClanListResponse{})
	}))
	defer server.Close()

	c := New(Config{BaseURL: server.URL, APIToken: "secret"})
	_, err := c.SearchClans(context.Background(), QuerySearchClans{
		Name:       "Test",
		MinMembers: 30,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("SearchClans error: %v", err)
	}
	// 查询参数应包含 name、minMembers、limit(零值的 after/before 不应出现)
	for _, want := range []string{"name=Test", "minMembers=30", "limit=20"} {
		if !containsParam(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if containsParam(gotQuery, "after=") || containsParam(gotQuery, "before=") {
		t.Fatalf("zero query params should be omitted: %q", gotQuery)
	}
}

func containsParam(query, param string) bool {
	parts := splitQuery(query)
	for _, p := range parts {
		if p == param {
			return true
		}
	}
	return false
}

func splitQuery(q string) []string {
	if q == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i < len(q); i++ {
		if q[i] == '&' {
			parts = append(parts, q[start:i])
			start = i + 1
		}
	}
	parts = append(parts, q[start:])
	return parts
}

func TestCheckStatusCodeMapsErrors(t *testing.T) {
	cases := []struct {
		code int
		want error
	}{
		{http.StatusOK, nil},
		{http.StatusUnauthorized, ErrAPIAccessDenied},
		{http.StatusForbidden, ErrAPIAccessDenied},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusTooManyRequests, ErrRateLimited},
	}
	for _, tc := range cases {
		if got := checkStatusCode(tc.code); !errors.Is(got, tc.want) {
			if tc.want == nil && got == nil {
				continue
			}
			t.Fatalf("status %d: got %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestNormalizeTag(t *testing.T) {
	cases := map[string]string{
		"2pp":   "#2PP",
		"#2PP":  "#2PP",
		" 2pp ": "#2PP",
		"":      "",
		"abc":   "#ABC",
	}
	for in, want := range cases {
		if got := NormalizeTag(in); got != want {
			t.Fatalf("NormalizeTag(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateTag(t *testing.T) {
	if _, err := ValidateTag("#2PP"); err != nil {
		t.Fatalf("valid tag error: %v", err)
	}
	if _, err := ValidateTag(""); err == nil {
		t.Fatal("expected error for empty tag")
	}
	if _, err := ValidateTag("#X"); err == nil {
		t.Fatal("expected error for too-short tag")
	}
}
