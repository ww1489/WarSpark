package errors

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	e := New(ErrCodeInvalidTag, "invalid tag")
	if e.Error() != "invalid tag" {
		t.Fatalf("got %q, want %q", e.Error(), "invalid tag")
	}
}

func TestError_ErrorWithWrapped(t *testing.T) {
	inner := errors.New("inner error")
	e := Wrap(ErrCodeAPIRequestFailed, "request failed", inner)
	got := e.Error()
	if got != "request failed: inner error" {
		t.Fatalf("got %q, want %q", got, "request failed: inner error")
	}
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	e := Wrap(ErrCodeAPIRequestFailed, "msg", inner)
	if !errors.Is(e, inner) {
		t.Fatal("Unwrap should allow errors.Is to reach inner error")
	}
}

func TestNew(t *testing.T) {
	e := New(ErrCodeClanNotFound, "clan not found")
	if e.Code != ErrCodeClanNotFound {
		t.Fatalf("code = %q, want %q", e.Code, ErrCodeClanNotFound)
	}
	if e.Message != "clan not found" {
		t.Fatalf("message = %q", e.Message)
	}
	if e.Err != nil {
		t.Fatal("New should not wrap an error")
	}
}

func TestWrap(t *testing.T) {
	inner := errors.New("inner")
	e := Wrap(ErrCodeAPIResponseInvalid, "invalid response", inner)
	if e.Code != ErrCodeAPIResponseInvalid {
		t.Fatalf("code = %q", e.Code)
	}
	if e.Message != "invalid response" {
		t.Fatalf("message = %q", e.Message)
	}
	if e.Err != inner {
		t.Fatal("Wrap should store the inner error")
	}
}

func TestCode(t *testing.T) {
	t.Run("domain error", func(t *testing.T) {
		e := New(ErrCodeWarNotFound, "not found")
		if got := Code(e); got != ErrCodeWarNotFound {
			t.Fatalf("got %q, want %q", got, ErrCodeWarNotFound)
		}
	})

	t.Run("wrapped domain error", func(t *testing.T) {
		e := Wrap(ErrCodePlayerNotFound, "not found", errors.New("api error"))
		if got := Code(e); got != ErrCodePlayerNotFound {
			t.Fatalf("got %q, want %q", got, ErrCodePlayerNotFound)
		}
	})

	t.Run("standard error", func(t *testing.T) {
		if got := Code(errors.New("plain error")); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if got := Code(nil); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}

func TestErrorCodeConstants(t *testing.T) {
	// Verify all constants have values
	codes := map[string]string{
		"invalid_tag":          ErrCodeInvalidTag,
		"api_not_configured":   ErrCodeAPINotConfigured,
		"api_request_failed":   ErrCodeAPIRequestFailed,
		"api_access_denied":    ErrCodeAPIAccessDenied,
		"war_not_found":        ErrCodeWarNotFound,
		"api_response_invalid": ErrCodeAPIResponseInvalid,
		"snapshot_not_found":   ErrCodeSnapshotNotFound,
		"clan_not_found":       ErrCodeClanNotFound,
		"player_not_found":     ErrCodePlayerNotFound,
		"league_not_found":     ErrCodeLeagueNotFound,
		"location_not_found":   ErrCodeLocationNotFound,
	}
	for expect, actual := range codes {
		if actual != expect {
			t.Fatalf("constant %q = %q", actual, expect)
		}
	}
}

func TestErrSnapshotNotFound(t *testing.T) {
	if ErrSnapshotNotFound == nil {
		t.Fatal("ErrSnapshotNotFound should not be nil")
	}
	if ErrSnapshotNotFound.Error() != "war snapshot not found" {
		t.Fatalf("got %q", ErrSnapshotNotFound.Error())
	}
}
