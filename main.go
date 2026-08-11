package main

import (
	"context"
	"log"
	"net/http"

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

		// Требует авторизации
		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)
			// auth
			r.Post("/logout", authHandler.Logout)

			// chat
			r.Route("/chats", func(r chi.Router) {
				r.Post("/", chatHandler.CreateChat)
				r.Get("/", chatHandler.GetChats)

				r.Route("/{chatID}", func(r chi.Router) {
					r.Route("/members", func(r chi.Router) {
						r.Post("/{userID}", chatHandler.AddChatMember)
						r.Get("/{userID}", chatHandler.GetUserChatMessage)
						r.Delete("/{userID}", chatHandler.DeleteChatMember)
					})
				})

				r.Route("/messages", func(r chi.Router) {
					r.Post("/", chatHandler.CreateMessage)
					r.Delete("/{messageID}", chatHandler.DeleteMessage)
					r.Patch("/{messageID}", chatHandler.EditMessage)
				})
			})
		})
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
