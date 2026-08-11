package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"messages/internal/models"
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
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}

	var chat models.Chat
	json.NewDecoder(r.Body).Decode(&chat)
	chatID, err := h.chatRepo.CreateChat(r.Context(), ownnerID, chat.Name)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"status":  "succes",
		"chat_id": chatID,
		"message": "chat created",
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
}

func (h *ChatHandler) AddChatMember(w http.ResponseWriter, r *http.Request) {
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
