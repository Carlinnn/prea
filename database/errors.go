package database

import "net/http"

// The driver's errors are translated into these. HTTPStatus gives core.Error
// the status to reply with, without core importing this package.
var (
	ErrNotFound     error = &statusError{"not found", http.StatusNotFound}       // no row
	ErrConflict     error = &statusError{"conflict", http.StatusConflict}        // unique violation
	ErrInvalidInput error = &statusError{"invalid input", http.StatusBadRequest} // malformed value, not null, foreign key, check
)

type statusError struct {
	msg    string
	status int
}

func (e *statusError) Error() string   { return e.msg }
func (e *statusError) HTTPStatus() int { return e.status }
