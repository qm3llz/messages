package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"messages/internal/db"
	"messages/internal/handler"
	"messages/internal/repository"
	"messages/internal/ws"

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

	hub := ws.NewHub()
	go hub.Run(ctx)

	wsHandler := handler.NewWSHandler(hub)
	authHandler := handler.NewAuthHandler(userRepo, sessionRepo)
	chatHandler := handler.NewChatHandler(chatRepo, hub)

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)
			r.Post("/logout", authHandler.Logout)
			r.Get("/ws", wsHandler.ServeWS)

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

	printRoutes(r)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}

func printRoutes(r chi.Router) {
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-7s %s\n", method, route)
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		log.Printf("Walk error: %s\n", err)
	}
}
