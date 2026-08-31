package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"messages/internal/db"
	"messages/internal/handler"
	"messages/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real env vars")
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx := context.Background()

	pool := db.ConnectDB(ctx)
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	chatRepo := repository.NewChatRepository(pool)

	authHandler := handler.NewAuthHandler(userRepo, sessionRepo)
	chatHandler := handler.NewChatHandler(chatRepo)

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)
			r.Post("/logout", authHandler.Logout)

			r.Route("/chats", func(r chi.Router) {
				r.Post("/", chatHandler.CreateChat)
				r.Get("/", chatHandler.GetUserChats)

				r.Route("/{chatID}", func(r chi.Router) {
					r.Route("/members", func(r chi.Router) {
						r.Get("/", chatHandler.GetChatMembers)
						r.Post("/{userID}", chatHandler.AddChatMember)
						r.Delete("/{userID}", chatHandler.RemoveChatMember)
					})

					r.Route("/messages", func(r chi.Router) {
						r.Get("/", chatHandler.GetChatMessages)
						r.Post("/", chatHandler.CreateMessage)
						r.Delete("/{messageID}", chatHandler.DeleteMessage)
						r.Patch("/{messageID}", chatHandler.UpdateMessage)
					})
				})
			})
		})
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
