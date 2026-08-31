package apperr

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrNotOwner      = errors.New("user not owner")

	ErrSessionNotFound = errors.New("session not found or expired")

	ErrMessageNotFoundOrForbidden = errors.New("message not found or access denied")

	ErrInvalidCredentials = errors.New("invalid username or password")

	ErrAlreadyMember = errors.New("user is already a member of this chat")

	ErrNotChatMember = errors.New("user is not a chat member")
)
