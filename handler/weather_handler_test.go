package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	// Ajuste o import path para o seu projeto, se necessário
	"github.com/MchlAlex/fc-lab01/internal/entity"
	"github.com/MchlAlex/fc-lab01/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLocationFinder é um mock para service.LocationFinder.
type MockLocationFinder struct {
	mock.Mock
}

func (m *MockLocationFinder) GetLocationByCEP(cep string) (string, error) {
	args := m.Called(cep)
	return args.String(0), args.Error(1)
}

// MockWeatherFinder é um mock para service.WeatherFinder.
type MockWeatherFinder struct {
	mock.Mock
}

func (m *MockWeatherFinder) GetWeatherByCity(city string) (float64, error) {
	args := m.Called(city)
	// Precisamos converter o primeiro argumento para float64
	// Adiciona verificação para evitar panic se Get(0) não for float64
	val, ok := args.Get(0).(float64)
	if !ok {
		// Se não for float64 (pode acontecer em cenários de erro onde o retorno é 0),
		// retorne um valor padrão ou trate conforme necessário.
		// Neste caso, retornar 0.0 é razoável se o erro for o foco.
		return 0.0, args.Error(1)
	}
	return val, args.Error(1)
}

// MockTemperatureConverter é um mock para service.TemperatureConverter.
type MockTemperatureConverter struct {
	mock.Mock
}

func (m *MockTemperatureConverter) ConvertTemperatures(tempC float64) *entity.WeatherOutput {
	args := m.Called(tempC)
	// Retorna o ponteiro para WeatherOutput ou nil se não for encontrado
	if output, ok := args.Get(0).(*entity.WeatherOutput); ok {
		return output
	}
	return nil
}

func TestWeatherHandler_GetWeatherByCEP(t *testing.T) {
	// Helper para configurar o roteador, evita repetição
	setupRouter := func(h *WeatherHandler) *chi.Mux {
		r := chi.NewRouter()
		r.Get("/weather/{cep}", h.GetWeatherByCEP)
		return r
	}

	t.Run("Success", func(t *testing.T) {
		// --- Criação dos mocks e handler DENTRO do t.Run ---
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)
		mockConverter := new(MockTemperatureConverter)
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)
		// --- Fim da criação ---

		cep := "01001000"
		city := "São Paulo"
		tempC := 25.0
		expectedOutput := &entity.WeatherOutput{TempC: 25.0, TempF: 77.0, TempK: 298.15}

		// Configura os mocks
		mockLocation.On("GetLocationByCEP", cep).Return(city, nil).Once()
		mockWeather.On("GetWeatherByCity", city).Return(tempC, nil).Once()
		mockConverter.On("ConvertTemperatures", tempC).Return(expectedOutput).Once()

		// Cria a requisição e o recorder
		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()

		// Executa o handler
		r.ServeHTTP(rr, req)

		// Verifica o resultado
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var actualOutput entity.WeatherOutput
		err := json.Unmarshal(rr.Body.Bytes(), &actualOutput)
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput.TempC, actualOutput.TempC)
		assert.InDelta(t, expectedOutput.TempF, actualOutput.TempF, 0.01) // Use InDelta for floats
		assert.InDelta(t, expectedOutput.TempK, actualOutput.TempK, 0.01)

		// Verifica se os mocks foram chamados
		mockLocation.AssertExpectations(t)
		mockWeather.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Invalid CEP Format", func(t *testing.T) {
		// --- Criação dos mocks e handler DENTRO do t.Run ---
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)          // Criado mas não configurado/chamado
		mockConverter := new(MockTemperatureConverter) // Criado mas não configurado/chamado
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)
		// --- Fim da criação ---

		cep := "12345" // CEP inválido

		// Configura o mock de localização para retornar erro de formato inválido
		mockLocation.On("GetLocationByCEP", cep).Return("", service.ErrInvalidCEPFormat).Once()
		// NÃO configurar .On() para mockWeather ou mockConverter

		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code) // 422

		var errorResp entity.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid zipcode", errorResp.Message)

		mockLocation.AssertExpectations(t)
		// Os outros mocks não devem ser chamados - AssertNotCalled agora funciona
		mockWeather.AssertNotCalled(t, "GetWeatherByCity", mock.Anything)
		mockConverter.AssertNotCalled(t, "ConvertTemperatures", mock.Anything)
	})

	t.Run("CEP Not Found", func(t *testing.T) {
		// --- Criação dos mocks e handler DENTRO do t.Run ---
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)
		mockConverter := new(MockTemperatureConverter)
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)
		// --- Fim da criação ---

		cep := "99999999" // CEP não encontrado

		// Configura o mock de localização para retornar erro de não encontrado
		mockLocation.On("GetLocationByCEP", cep).Return("", service.ErrCEPNotFound).Once()

		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code) // 404

		var errorResp entity.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "can not find zipcode", errorResp.Message)

		mockLocation.AssertExpectations(t)
		mockWeather.AssertNotCalled(t, "GetWeatherByCity", mock.Anything)
		mockConverter.AssertNotCalled(t, "ConvertTemperatures", mock.Anything)
	})

	t.Run("Location Service Internal Error", func(t *testing.T) {
		// --- Criação dos mocks e handler DENTRO do t.Run ---
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)
		mockConverter := new(MockTemperatureConverter)
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)
		// --- Fim da criação ---

		cep := "01001000"
		internalError := errors.New("some internal ViaCEP error")

		// Configura o mock de localização para retornar um erro genérico
		mockLocation.On("GetLocationByCEP", cep).Return("", internalError).Once()

		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code) // 500
		// Verifica se a mensagem de erro padrão é retornada (ou pode verificar o log)
		assert.Contains(t, rr.Body.String(), "Internal server error while fetching location")

		mockLocation.AssertExpectations(t)
		mockWeather.AssertNotCalled(t, "GetWeatherByCity", mock.Anything)
		mockConverter.AssertNotCalled(t, "ConvertTemperatures", mock.Anything)
	})

	t.Run("Weather Service Error", func(t *testing.T) {
		// --- Criação dos mocks e handler DENTRO do t.Run ---
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)
		mockConverter := new(MockTemperatureConverter)
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)
		// --- Fim da criação ---

		cep := "01001000"
		city := "São Paulo"
		weatherError := errors.New("weather API unavailable")

		// Configura os mocks
		mockLocation.On("GetLocationByCEP", cep).Return(city, nil).Once()
		// Passa 0.0 explicitamente como float64 para o retorno do mock
		mockWeather.On("GetWeatherByCity", city).Return(0.0, weatherError).Once()

		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code) // 500
		assert.Contains(t, rr.Body.String(), "Internal server error while fetching weather data")

		mockLocation.AssertExpectations(t)
		mockWeather.AssertExpectations(t)
		// O conversor não deve ser chamado
		mockConverter.AssertNotCalled(t, "ConvertTemperatures", mock.Anything)
	})
}
