package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MchlAlex/fc-lab02/config"
	"github.com/MchlAlex/fc-lab02/internal/infra/web"
)

func main() {
	// Carrega a configuração
	cfg, err := config.LoadConfig(".") // "." indica o diretório atual
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	// Configura o servidor web
	router := web.SetupServer(cfg)

	// Define a porta do servidor
	port := cfg.WebServerPort
	if port == "" {
		port = "8080" // Porta padrão se não definida
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting server on port %s", port)

	// Inicia o servidor HTTP
	err = http.ListenAndServe(addr, router)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", addr, err)
	}

	log.Println("Server stopped")
}
