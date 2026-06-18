package coc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appconfig "github.com/ww1489/WarSpark/internal/config"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

func TestClientCurrentWarSendsBearerTokenAndEscapedTag(t *testing.T) {
	var gotPath string
	var gotAuthorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotAuthorization = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(wardomain.CurrentWar{
			State:    "inWar",
			TeamSize: 15,
			Clan:     wardomain.WarClan{Tag: "#AAA111"},
			Opponent: wardomain.WarClan{Tag: "#BBB222"},
		})
	}))
	defer server.Close()

	client := New(appconfig.CoCConfig{
		BaseURL:  server.URL,
		APIToken: "secret-token",
		Timeout:  time.Second,
	})

	currentWar, err := client.CurrentWar(context.Background(), "#AAA111")
	if err != nil {
		t.Fatalf("CurrentWar returned error: %v", err)
	}

	if gotPath != "/clans/%23AAA111/currentwar" {
		t.Fatalf("expected escaped tag path, got %q", gotPath)
	}
	if gotAuthorization != "Bearer secret-token" {
		t.Fatalf("expected bearer token, got %q", gotAuthorization)
	}
	if currentWar.Opponent.Tag != "#BBB222" {
		t.Fatalf("unexpected response: %#v", currentWar)
	}
}

func TestClientCurrentWarRequiresToken(t *testing.T) {
	client := New(appconfig.CoCConfig{
		BaseURL: "https://api.clashofclans.com/v1",
		Timeout: time.Second,
	})

	_, err := client.CurrentWar(context.Background(), "#AAA111")
	if err == nil {
		t.Fatal("expected api_not_configured error")
	}
	if got := wardomain.ErrorCode(err); got != wardomain.ErrorAPINotConfigured {
		t.Fatalf("expected api_not_configured, got %q", got)
	}
}
