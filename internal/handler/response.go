package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"messages/internal/apperr"
)

var errNoUserInContext = errors.New("userID missing from request context")
var errInvalidParam = errors.New("invalid path parameter")

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotOwner):
		writeResponse(w, http.StatusForbidden, errorResponse{"not_chat_owner", "only chat owner can do that"})

	case errors.Is(err, apperr.ErrUsernameTaken):
		writeResponse(w, http.StatusConflict, errorResponse{"username_taken", "username already taken"})

	case errors.Is(err, apperr.ErrUserNotFound) || errors.Is(err, apperr.ErrInvalidCredentials):
		writeResponse(w, http.StatusUnauthorized, errorResponse{"pass_or_user_invalid", "invalid password or username"})

	case errors.Is(err, apperr.ErrSessionNotFound):
		writeResponse(w, http.StatusUnauthorized, errorResponse{"unauthorized", "session not found, login again"})

	case errors.Is(err, apperr.ErrMessageNotFoundOrForbidden):
		writeResponse(w, http.StatusNotFound, errorResponse{"message_not_found", "message not found or forbidden"})

	case errors.Is(err, errInvalidParam):
		writeResponse(w, http.StatusBadRequest, errorResponse{"invalid_parameter", "invalid path parameter"})

	default:
		slog.Error("server_error", "err", err)
		writeResponse(w, http.StatusInternalServerError, errorResponse{"server_error", "server error"})
	}
}

// writeResponse ??? // TODO: docs
func writeResponse(w http.ResponseWriter, status int, v any) {
	if status == http.StatusNoContent && v == nil {
		w.WriteHeader(status)
		return
	}

	buf, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(buf)
	if err != nil {
		// TODO: log
	}
}
