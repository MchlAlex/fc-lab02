package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/MchlAlex/fc-lab01/internal/entity"
	"github.com/MchlAlex/fc-lab01/internal/service"

	"github.com/go-chi/chi/v5"
)

// WeatherHandler contém as dependências para o handler de clima.
type WeatherHandler struct {
	LocationService service.LocationFinder
	WeatherService  service.WeatherFinder
	Converter       service.TemperatureConverter
}

// NewWeatherHandler cria uma nova instância de WeatherHandler.
func NewWeatherHandler(loc service.LocationFinder, weather service.WeatherFinder, conv service.TemperatureConverter) *WeatherHandler {
	return &WeatherHandler{
		LocationService: loc,
		WeatherService:  weather,
		Converter:       conv,
	}
}

// GetWeatherByCEP é o handler para a rota GET /weather/{cep}.
func (h *WeatherHandler) GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")
	if cep == "" {
		http.Error(w, "CEP parameter is missing", http.StatusBadRequest)
		return
	}

	// 1. Buscar localização pelo CEP
	city, err := h.LocationService.GetLocationByCEP(cep)
	if err != nil {
		log.Printf("Error finding location for CEP %s: %v", cep, err)
		if errors.Is(err, service.ErrInvalidCEPFormat) {
			w.WriteHeader(http.StatusUnprocessableEntity) // 422
			json.NewEncoder(w).Encode(entity.ErrorResponse{Message: "invalid zipcode"})
			return
		}
		if errors.Is(err, service.ErrCEPNotFound) {
			w.WriteHeader(http.StatusNotFound) // 404
			json.NewEncoder(w).Encode(entity.ErrorResponse{Message: "can not find zipcode"})
			return
		}
		// Outros erros (falha na API ViaCEP, etc.)
		http.Error(w, "Internal server error while fetching location", http.StatusInternalServerError)
		return
	}

	// 2. Buscar clima pela cidade
	tempC, err := h.WeatherService.GetWeatherByCity(city)
	if err != nil {
		log.Printf("Error finding weather for city %s (from CEP %s): %v", city, cep, err)
		// Tratar erros específicos da WeatherAPI se necessário, mas por padrão retorna 500
		http.Error(w, "Internal server error while fetching weather data", http.StatusInternalServerError)
		return
	}

	// 3. Converter temperaturas
	weatherOutput := h.Converter.ConvertTemperatures(tempC)

	// 4. Responder com sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(weatherOutput)
}
