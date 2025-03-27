package service

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	// Ajuste o import path se necessário
	"github.com/MchlAlex/fc-lab02/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoundTripper é um mock para http.RoundTripper usado nos testes.
type MockRoundTripper struct {
	mock.Mock
}

// RoundTrip implementa a interface http.RoundTripper.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	// Verifica se o primeiro argumento é um *http.Response antes de tentar acessá-lo
	if resp, ok := args.Get(0).(*http.Response); ok {
		return resp, args.Error(1)
	}
	// Se não for, retorna nil para a resposta e o erro configurado
	return nil, args.Error(1)
}

func TestViaCEPService_GetLocationByCEP(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		viaCEPService := NewViaCEPService(mockClient)
		// --- Fim da criação ---

		cep := "01001000"
		expectedCity := "São Paulo"
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"localidade": "São Paulo", "uf": "SP"}`)),
			Header:     make(http.Header),
		}
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

		city, err := viaCEPService.GetLocationByCEP(cep)

		assert.NoError(t, err)
		assert.Equal(t, expectedCity, city)
		mockTripper.AssertExpectations(t)
	})

	t.Run("Invalid CEP Format", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper) // Criado mas não configurado/chamado
		mockClient := &http.Client{Transport: mockTripper}
		viaCEPService := NewViaCEPService(mockClient)
		// --- Fim da criação ---

		cep := "12345" // Formato inválido
		city, err := viaCEPService.GetLocationByCEP(cep)

		assert.ErrorIs(t, err, ErrInvalidCEPFormat)
		assert.Empty(t, city)
		// Não deve fazer chamada HTTP - AssertNotCalled agora funciona
		mockTripper.AssertNotCalled(t, "RoundTrip", mock.AnythingOfType("*http.Request"))
	})

	t.Run("CEP Not Found (API Error)", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		viaCEPService := NewViaCEPService(mockClient)
		// --- Fim da criação ---

		cep := "99999999" // CEP inexistente
		mockResponse := &http.Response{
			StatusCode: http.StatusOK, // ViaCEP retorna 200 mesmo com erro
			Body:       io.NopCloser(bytes.NewBufferString(`{"erro": true}`)),
			Header:     make(http.Header),
		}
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

		city, err := viaCEPService.GetLocationByCEP(cep)

		assert.ErrorIs(t, err, ErrCEPNotFound)
		assert.Empty(t, city)
		mockTripper.AssertExpectations(t)
	})

	t.Run("ViaCEP API Failure", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		viaCEPService := NewViaCEPService(mockClient)
		// --- Fim da criação ---

		cep := "01001000"
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("network error")).Once()

		city, err := viaCEPService.GetLocationByCEP(cep)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute ViaCEP request")
		assert.Empty(t, city)
		mockTripper.AssertExpectations(t)
	})

	t.Run("ViaCEP Non-OK Status", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		viaCEPService := NewViaCEPService(mockClient)
		// --- Fim da criação ---

		cep := "01001000"
		mockResponse := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("server error")),
			Header:     make(http.Header),
		}
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

		city, err := viaCEPService.GetLocationByCEP(cep)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ViaCEP request failed with status 500")
		assert.Empty(t, city)
		mockTripper.AssertExpectations(t)
	})
}

func TestWeatherAPIService_GetWeatherByCity(t *testing.T) {
	// apiKey pode ficar fora se for constante entre os testes
	apiKey := "test-api-key"

	t.Run("Success", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		weatherService := NewWeatherAPIService(apiKey, mockClient)
		// --- Fim da criação ---

		city := "São Paulo"
		expectedTempC := 25.5
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"current": {"temp_c": 25.5}}`)),
			Header:     make(http.Header),
		}
		// Verifica se a URL contém a cidade encodada e a API key
		mockTripper.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "api.weatherapi.com" &&
				req.URL.Path == "/v1/current.json" &&
				req.URL.Query().Get("key") == apiKey &&
				req.URL.Query().Get("q") == city // QueryEscape é testado implicitamente
		})).Return(mockResponse, nil).Once()

		tempC, err := weatherService.GetWeatherByCity(city)

		assert.NoError(t, err)
		assert.Equal(t, expectedTempC, tempC)
		mockTripper.AssertExpectations(t)
	})

	t.Run("WeatherAPI Failure", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		weatherService := NewWeatherAPIService(apiKey, mockClient)
		// --- Fim da criação ---

		city := "London"
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("network error")).Once()

		tempC, err := weatherService.GetWeatherByCity(city)

		assert.ErrorIs(t, err, ErrWeatherAPIFailure)
		assert.Contains(t, err.Error(), "network error")
		assert.Zero(t, tempC)
		mockTripper.AssertExpectations(t)
	})

	t.Run("WeatherAPI Non-OK Status", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		weatherService := NewWeatherAPIService(apiKey, mockClient)
		// --- Fim da criação ---

		city := "InvalidCity"
		mockResponse := &http.Response{
			StatusCode: http.StatusBadRequest, // Exemplo de erro da API
			Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"code": 1006, "message": "No matching location found."}}`)),
			Header:     make(http.Header),
		}
		mockTripper.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

		tempC, err := weatherService.GetWeatherByCity(city)

		assert.ErrorIs(t, err, ErrWeatherAPIFailure)
		assert.Contains(t, err.Error(), "status 400 - No matching location found.")
		assert.Zero(t, tempC)
		mockTripper.AssertExpectations(t)
	})

	t.Run("Missing API Key", func(t *testing.T) {
		// --- Criação DENTRO do t.Run ---
		// Criamos o mockTripper e mockClient aqui também, mesmo que não devam ser usados
		mockTripper := new(MockRoundTripper)
		mockClient := &http.Client{Transport: mockTripper}
		// O serviço é criado com a chave vazia
		weatherServiceNoKey := NewWeatherAPIService("", mockClient)
		// --- Fim da criação ---

		city := "São Paulo"

		tempC, err := weatherServiceNoKey.GetWeatherByCity(city)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WeatherAPI key is missing")
		assert.Zero(t, tempC)
		// Verifica que nenhuma chamada HTTP foi feita - AssertNotCalled agora funciona
		mockTripper.AssertNotCalled(t, "RoundTrip", mock.Anything)
	})
}

// TestStandardTemperatureConverter_ConvertTemperatures não usa mocks que precisam ser resetados,
// então pode permanecer como está.
func TestStandardTemperatureConverter_ConvertTemperatures(t *testing.T) {
	converter := NewStandardTemperatureConverter()

	tests := []struct {
		name     string
		tempC    float64
		expected entity.WeatherOutput
	}{
		{"Zero Celsius", 0, entity.WeatherOutput{TempC: 0, TempF: 32, TempK: 273.15}},
		{"Positive Celsius", 25, entity.WeatherOutput{TempC: 25, TempF: 77, TempK: 298.15}},
		{"Negative Celsius", -10, entity.WeatherOutput{TempC: -10, TempF: 14, TempK: 263.15}},
		{"Decimal Celsius", 15.5, entity.WeatherOutput{TempC: 15.5, TempF: 59.9, TempK: 288.65}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertTemperatures(tt.tempC)
			// Usar assert.InDelta para comparações de float
			assert.InDelta(t, tt.expected.TempC, result.TempC, 0.01)
			assert.InDelta(t, tt.expected.TempF, result.TempF, 0.01)
			assert.InDelta(t, tt.expected.TempK, result.TempK, 0.01)
		})
	}
}
