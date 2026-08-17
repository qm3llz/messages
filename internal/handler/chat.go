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
	GetChatMessages(ctx context.Context, chatID, beforeID, requesterID, limit int) ([]models.Message, error)
	GetChatMembers(ctx context.Context, chatID, requesterID int) ([]models.User, error)
	CreateMessage(ctx context.Context, chatID, requesterID int, text string) error
	UpdateMessage(ctx context.Context, newText string, messageID, requesterID int) error
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
	writeJSON(w, http.StatusCreated, response)
}

func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) { // TODO: проверить все
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

	writeJSON(w, http.StatusOK, chats)
}

func (h *ChatHandler) AddChatMember(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(chi.URLParam(r, "chatID"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatHandler) GetUserChatMessage(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) DeleteChatMember(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) EditMessage(w http.ResponseWriter, r *http.Request) {}
