package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed frontend
var frontendFS embed.FS

type statusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

const engineVersion = "1.27.1"

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	trajectoryPath := flag.String("trajectory", "output/trajectory.json", "Path to trajectory JSON file")
	flag.Parse()

	fmt.Println("🦅 BLACK SWAN EVENT (BSE) VISUALIZER")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("[🌐] Starting local offline web server on http://localhost:%s\n", *port)
	fmt.Printf("[📂] Serving trajectory from: %s\n", *trajectoryPath)

	sub, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		fmt.Printf("[❌] Failed to load embedded frontend: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))

	mux.HandleFunc("/api/simulation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := statusResponse{
			Status:  "BSE physics engine ready.",
			Version: engineVersion,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/trajectory", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(*trajectoryPath)
		if err != nil {
			http.Error(w, "trajectory not available: run the CLI first", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if _, err := w.Write(data); err != nil {
			return
		}
	})

	server := &http.Server{
		Addr:              ":" + *port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[❌] Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println("[🔔] Server is running.")
	fmt.Println("[💡] Press Ctrl+C to stop the server.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println()
	fmt.Println("[🛑] Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("[❌] Forced shutdown: %v\n", err)
	}
}
