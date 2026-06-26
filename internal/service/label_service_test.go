package service

import (
	"context"
	"testing"

	"github.com/ww1489/WarSpark/internal/domain/label"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeLabelCache struct {
	clanLabels   label.LabelListResponse
	clanHit      bool
	playerLabels label.LabelListResponse
	playerHit    bool
}

func (f *fakeLabelCache) GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	return f.clanLabels, f.clanHit, nil
}
func (f *fakeLabelCache) SetClanLabels(ctx context.Context, resp label.LabelListResponse) error {
	return nil
}
func (f *fakeLabelCache) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	return f.playerLabels, f.playerHit, nil
}
func (f *fakeLabelCache) SetPlayerLabels(ctx context.Context, resp label.LabelListResponse) error {
	return nil
}

func TestLabelServiceGetClanLabelsCached(t *testing.T) {
	cached := label.LabelListResponse{Items: []label.Label{{ID: 1, Name: "Clan War"}}}
	svc := NewLabelService(&fakeCocapiClient{}, &fakeLabelCache{clanLabels: cached, clanHit: true})
	resp, err := svc.GetClanLabels(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Clan War" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLabelServiceGetClanLabelsFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{clanLabels: cocapi.LabelListResponse{Items: []cocapi.Label{{ID: 1, Name: "Clan War"}}}}
	svc := NewLabelService(fakeAPI, &fakeLabelCache{})
	resp, err := svc.GetClanLabels(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Clan War" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLabelServiceGetPlayerLabels(t *testing.T) {
	fakeAPI := &fakeCocapiClient{playerLabels: cocapi.LabelListResponse{Items: []cocapi.Label{{ID: 2, Name: "Active"}}}}
	svc := NewLabelService(fakeAPI, &fakeLabelCache{})
	resp, err := svc.GetPlayerLabels(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Active" {
		t.Fatalf("got %+v", resp.Items)
	}
}
