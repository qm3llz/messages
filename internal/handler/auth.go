package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"messages/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type userRepo interface {
	CreateUser(ctx context.Context, username, passwordHash string) (int, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdatePasswordHash(ctx context.Context, passwordHash, username string) error
}

type sessionRepo interface {
	CreateSession(ctx context.Context, token string, userID int, expiresAt time.Time) error
	GetUserIDByToken(ctx context.Context, token string) (int, error)
	DeleteSessionByToken(ctx context.Context, token string) error
}

type AuthHandler struct {
	userRepo    userRepo
	sessionRepo sessionRepo
}

func NewAuthHandler(userRepo userRepo, sessionRepository sessionRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, sessionRepo: sessionRepository}
}

func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h *AuthHandler) startSession(ctx context.Context, w http.ResponseWriter, userID int) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err // TODO: Нету коллизии для рандома
	}

	token := base64.RawURLEncoding.EncodeToString(b)
	hashedToken := hashToken(token)

	expiresAt := time.Now().AddDate(0, 0, 7)

	err := h.sessionRepo.CreateSession(ctx, hashedToken, userID, expiresAt)
	if err != nil {
		return err // TODO: log + generic message
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var u models.UserLogin

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		writeError(w, err)
		return
	}

	if len(u.Username) < 5 || len(u.Username) > 64 {
		writeResponse(w, http.StatusBadRequest, errorResponse{Code: "bad_username", Message: "The username must contain from 5 to 64 characters"})
		return
	}
	if len(u.Password) < 8 || len(u.Password) > 72 {
		writeResponse(w, http.StatusBadRequest, errorResponse{Code: "bad_password", Message: "The password must contain from 8 to 72 characters"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, err)
		return
	}

	userID, err := h.userRepo.CreateUser(r.Context(), u.Username, string(hashedPassword))
	if err != nil {
		writeError(w, err)
		return
	}

	err = h.startSession(r.Context(), w, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	response := map[string]any{
		"message":  "success",
		"userID":   userID,
		"username": u.Username,
	}
	writeResponse(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var u models.UserLogin

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		writeError(w, err)
		return
	}

	user, err := h.userRepo.GetUserByUsername(r.Context(), u.Username)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(u.Password)); err != nil {
		writeError(w, err)
		return
	}

	if err := h.startSession(r.Context(), w, user.ID); err != nil {
		writeError(w, err)
		return
	}

	response := map[string]any{
		"message":  "authorized",
		"userID":   user.ID,
		"username": u.Username,
	}
	writeResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		writeError(w, err)
		return
	}

	hashedToken := hashToken(session.Value)
	err = h.sessionRepo.DeleteSessionByToken(r.Context(), hashedToken)
	if err != nil {
		writeError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	response := map[string]any{
		"message": "logged out",
	}
	writeResponse(w, http.StatusOK, response)
}
