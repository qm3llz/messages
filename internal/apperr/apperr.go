package apperr

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrNotOwner      = errors.New("user not owner")

	ErrSessionNotFound = errors.New("session not found or expired")

	ErrMessageNotFoundOrForbidden = errors.New("message not found or access denied")
)
