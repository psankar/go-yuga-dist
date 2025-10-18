package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-yuga-dist/app/handlers"
	"go-yuga-dist/app/store"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Database connection
	dbp, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbp.Close()

	// HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: setupRoutes(dbp),
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
		}
	}()
	log.Printf("Server is ready to handle requests at %s", server.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRoutes(dbpool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	store := store.NewDB(dbpool)
	server := handlers.NewServer(store)

	mux.HandleFunc("/signup", server.Signup)
	mux.HandleFunc("/signin", server.Signin)

	// Authenticated routes
	mux.Handle("/user-posts", handlers.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.CreatePost(w, r)
		} else if r.Method == http.MethodGet {
			server.GetUserPosts(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}), store))
	mux.Handle("/user-post/", handlers.AuthMiddleware(http.HandlerFunc(server.GetUserPost), store))

	return mux
}
