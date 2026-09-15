package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"messages/internal/db"
	"messages/internal/handler"
	"messages/internal/repository"
	"messages/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real env vars")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
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
		r.Use(middleware.ClientIPFromRemoteAddr)
		r.Use(middleware.Logger)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		}) // TODO: Документацию по апи сюда

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitBy(10, time.Minute, func(r *http.Request) (string, error) {
				return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
			}))
			r.Post("/signup", authHandler.SignUp)
			r.Post("/login", authHandler.Login)
		})

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

	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}

	log.Printf("сервер запущен на %s\n", ln.Addr())

	if err := http.Serve(ln, r); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
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
