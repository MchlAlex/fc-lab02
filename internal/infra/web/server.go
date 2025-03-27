package web

import (
	"fmt"
	"net/http"

	"github.com/MchlAlex/fc-lab02/config"
	"github.com/MchlAlex/fc-lab02/handler"
	"github.com/MchlAlex/fc-lab02/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupServer configura e retorna o roteador HTTP.
func SetupServer(cfg *config.Config) *chi.Mux {
	// Inicializa os serviços com suas dependências
	locationService := service.NewViaCEPService(nil)                       // Usa http.DefaultClient
	weatherService := service.NewWeatherAPIService(cfg.WeatherAPIKey, nil) // Usa http.DefaultClient
	converter := service.NewStandardTemperatureConverter()

	// Inicializa o handler com os serviços
	weatherHandler := handler.NewWeatherHandler(locationService, weatherService, converter)

	// Configura o roteador Chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)    // Log das requisições
	r.Use(middleware.Recoverer) // Recupera de panics

	// Define a rota principal
	r.Get("/weather/{cep}", weatherHandler.GetWeatherByCEP)

	// Rota de health check (opcional, mas boa prática)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	return r
}
