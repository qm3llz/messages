package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"messages/internal/models"

	"github.com/go-chi/chi/v5"
)

type ChatRepo interface {
	CreateChat(ctx context.Context, ownerID int, name string) (int, error)
	AddChatMember(ctx context.Context, chatID, userID, requesterID int) error
	RemoveChatMember(ctx context.Context, chatID, userID, requesterID int) error
	GetUserChats(ctx context.Context, userID int) ([]models.Chat, error)
	GetChatMessages(ctx context.Context, chatID, requesterID int, f models.MessageFilter) ([]models.Message, error)
	GetChatMembers(ctx context.Context, chatID, requesterID int) ([]models.User, error)
	CreateMessage(ctx context.Context, chatID, requesterID int, text string) error
	UpdateMessage(ctx context.Context, newText string, messageID, requesterID int) error
	DeleteMessage(ctx context.Context, messageID, requesterID int) error
}

type ChatHandler struct {
	chatRepo ChatRepo
}

func NewChatHandler(chatRepo ChatRepo) *ChatHandler {
	return &ChatHandler{chatRepo: chatRepo}
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	ownnerID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	var chat models.Chat
	json.NewDecoder(r.Body).Decode(&chat)
	chatID, err := h.chatRepo.CreateChat(r.Context(), ownnerID, chat.Name)
	if err != nil {
		writeError(w, err)
		return
	}

	response := map[string]any{
		"chat_id": chatID,
	}
	writeResponse(w, http.StatusCreated, response)
}

func (h *ChatHandler) AddChatMember(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(chi.URLParam(r, "chatID"))
	if err != nil {
		writeError(w, errInvalidParam)
		return
	}

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, errInvalidParam)
		return
	}

	userReq, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	err = h.chatRepo.AddChatMember(r.Context(), chatID, userID, userReq)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, http.StatusNoContent, nil)
}

func (h *ChatHandler) RemoveChatMember(w http.ResponseWriter, r *http.Request) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	chatID, err := strconv.Atoi(chi.URLParam(r, "chatID"))
	if err != nil {
		writeError(w, errInvalidParam)
		return
	}
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, errInvalidParam)
		return
	}

	err = h.chatRepo.RemoveChatMember(r.Context(), chatID, userID, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, http.StatusNoContent, nil)
}

func (h *ChatHandler) GetUserChats(w http.ResponseWriter, r *http.Request) { // TODO: проверить все
	ownerID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	chats, err := h.chatRepo.GetUserChats(r.Context(), ownerID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, http.StatusOK, chats)
}

func (h *ChatHandler) GetChatMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	chatID, err := strconv.Atoi(chi.URLParam(r, "chatID"))
	if err != nil {
		writeError(w, errInvalidParam)
		return
	}

	q := r.URL.Query()
	params := models.MessageFilter{
		BeforeID: nil,
		UserID:   nil,
		Limit:    20,
	}

	if s := q.Get("before"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			writeError(w, errInvalidParam)
			return
		}
		params.BeforeID = &n
	}

	if s := q.Get("user"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			writeError(w, errInvalidParam)
			return
		}
		params.UserID = &n
	}

	if s := q.Get("limit"); s != "" {
		params.Limit, err = strconv.Atoi(s)
		if err != nil {
			writeError(w, errInvalidParam)
			return
		}
	}

	messages, err := h.chatRepo.GetChatMessages(r.Context(), chatID, id, params)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, http.StatusOK, messages)
}

func (h *ChatHandler) GetChatMembers(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
}
