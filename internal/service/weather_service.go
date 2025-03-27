package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/MchlAlex/fc-lab02/internal/entity"
)

// WeatherFinder define a interface para buscar o clima por cidade.
type WeatherFinder interface {
	GetWeatherByCity(city string) (float64, error)
}

// TemperatureConverter define a interface para converter temperaturas.
type TemperatureConverter interface {
	ConvertTemperatures(tempC float64) *entity.WeatherOutput
}

// WeatherAPIService implementa WeatherFinder usando a WeatherAPI.
type WeatherAPIService struct {
	APIKey string
	Client *http.Client
}

// NewWeatherAPIService cria uma nova instância de WeatherAPIService.
func NewWeatherAPIService(apiKey string, client *http.Client) *WeatherAPIService {
	if client == nil {
		client = http.DefaultClient
	}
	return &WeatherAPIService{APIKey: apiKey, Client: client}
}

var ErrWeatherAPIFailure = errors.New("failed to get weather data")

// GetWeatherByCity busca a temperatura atual (Celsius) para uma cidade usando a WeatherAPI.
func (s *WeatherAPIService) GetWeatherByCity(city string) (float64, error) {
	if s.APIKey == "" {
		return 0, errors.New("WeatherAPI key is missing")
	}

	// URL Encode a cidade para evitar problemas com espaços ou caracteres especiais
	encodedCity := url.QueryEscape(city)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.APIKey, encodedCity)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create WeatherAPI request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrWeatherAPIFailure, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read WeatherAPI response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Tenta decodificar uma possível mensagem de erro da API
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		errMsg := fmt.Sprintf("status %d", resp.StatusCode)
		if errorData, ok := errorResp["error"].(map[string]interface{}); ok {
			if msg, ok := errorData["message"].(string); ok {
				errMsg = fmt.Sprintf("status %d - %s", resp.StatusCode, msg)
			}
		}
		return 0, fmt.Errorf("%w: request failed with %s", ErrWeatherAPIFailure, errMsg)
	}

	var weatherResp entity.WeatherAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return 0, fmt.Errorf("failed to decode WeatherAPI response: %w", err)
	}

	return weatherResp.Current.TempC, nil
}

// StandardTemperatureConverter implementa TemperatureConverter.
type StandardTemperatureConverter struct{}

// NewStandardTemperatureConverter cria uma nova instância de StandardTemperatureConverter.
func NewStandardTemperatureConverter() *StandardTemperatureConverter {
	return &StandardTemperatureConverter{}
}

// ConvertTemperatures converte Celsius para Fahrenheit e Kelvin.
func (c *StandardTemperatureConverter) ConvertTemperatures(tempC float64) *entity.WeatherOutput {
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15 // Usando 273.15 para Kelvin, mais preciso que 273
	return &entity.WeatherOutput{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}
}
