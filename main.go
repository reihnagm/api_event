package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"superapps/controllers"
	helper "superapps/helpers"
	middleware "superapps/middlewares"
	"superapps/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Init DB (pakai yang sudah ada di project kamu)
	services.InitDBs()

	router := mux.NewRouter()

	// ✅ Middlewares
	router.Use(middleware.CorsMiddleware)
	router.Use(middleware.JwtAuthentication)

	rl := middleware.NewRateLimiterFromEnv()
	router.Use(rl.Middleware)

	// --- STATIC ---
	if err := os.MkdirAll("public", os.ModePerm); err != nil {
		log.Fatalf("Failed to create or access directory: %v", err)
	}

	dir, err := os.Open("public")
	if err != nil {
		log.Fatalf("Failed to open public directory: %v", err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatalf("Failed to read directory contents: %v", err)
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			staticPath := "/" + fileInfo.Name() + "/"
			publicPath := "./public/" + fileInfo.Name() + "/"
			log.Printf("Serving static files from %s at %s", publicPath, staticPath)
			router.PathPrefix(staticPath).
				Handler(http.StripPrefix(staticPath, http.FileServer(http.Dir(publicPath))))
		}
	}

	router.HandleFunc("/api/v1/auth/register", controllers.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/login", controllers.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/logout", controllers.Logout).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/v1/event/list", controllers.GetListEvent).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/event/detail", controllers.GetDetailEvent).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/event/create", controllers.CreateEvent).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/event/update", controllers.UpdateEvent).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/event/delete", controllers.DeleteEvent).Methods("DELETE", "OPTIONS")

	// --- SERVER ---
	port := ":" + firstNonEmpty(os.Getenv("PORT"), "9000")
	helper.Logger("info", "Starting server at "+port)

	server := &http.Server{
		Addr:              port,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	// Run server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			helper.Logger("error", fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)
	helper.Logger("info", "Server stopped gracefully")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
