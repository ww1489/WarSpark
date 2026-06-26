package errors

import "errors"

const (
	ErrCodeInvalidTag         = "invalid_tag"
	ErrCodeAPINotConfigured   = "api_not_configured"
	ErrCodeAPIRequestFailed   = "api_request_failed"
	ErrCodeAPIAccessDenied    = "api_access_denied"
	ErrCodeWarNotFound        = "war_not_found"
	ErrCodeAPIResponseInvalid = "api_response_invalid"
	ErrCodeSnapshotNotFound   = "snapshot_not_found"
	ErrCodeClanNotFound       = "clan_not_found"
	ErrCodePlayerNotFound     = "player_not_found"
	ErrCodeLeagueNotFound     = "league_not_found"
	ErrCodeLocationNotFound   = "location_not_found"
)

var ErrSnapshotNotFound = errors.New("war snapshot not found")

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Err
}

func New(code string, message string) Error {
	return Error{Code: code, Message: message}
}

func Wrap(code string, message string, err error) Error {
	return Error{Code: code, Message: message, Err: err}
}

func Code(err error) string {
	var appErr Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}
