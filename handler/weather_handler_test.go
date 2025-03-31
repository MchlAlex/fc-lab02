package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	// Ajuste o import path para o seu projeto, se necessário
	"github.com/MchlAlex/fc-lab02/internal/entity"

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
	setupRouter := func(h *WeatherHandler) *chi.Mux {
		r := chi.NewRouter()
		r.Get("/weather/{cep}", h.GetWeatherByCEP)
		return r
	}

	t.Run("Success", func(t *testing.T) {
		mockLocation := new(MockLocationFinder)
		mockWeather := new(MockWeatherFinder)
		mockConverter := new(MockTemperatureConverter)
		handler := NewWeatherHandler(mockLocation, mockWeather, mockConverter)
		r := setupRouter(handler)

		cep := "01001000"
		city := "São Paulo"
		tempC := 25.0
		// ✅ Inclui o campo "City" no expectedOutput
		expectedOutput := &entity.WeatherOutput{
			City:  city, // Novo campo
			TempC: 25.0,
			TempF: 77.0,
			TempK: 298.15,
		}

		mockLocation.On("GetLocationByCEP", cep).Return(city, nil).Once()
		mockWeather.On("GetWeatherByCity", city).Return(tempC, nil).Once()
		// ✅ O mockConverter deve retornar o expectedOutput completo
		mockConverter.On("ConvertTemperatures", tempC).Return(expectedOutput).Once()

		req := httptest.NewRequest("GET", "/weather/"+cep, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var actualOutput entity.WeatherOutput
		err := json.Unmarshal(rr.Body.Bytes(), &actualOutput)
		assert.NoError(t, err)

		// ✅ Valida o campo "City"
		assert.Equal(t, expectedOutput.City, actualOutput.City)
		assert.Equal(t, expectedOutput.TempC, actualOutput.TempC)
		assert.InDelta(t, expectedOutput.TempF, actualOutput.TempF, 0.01)
		assert.InDelta(t, expectedOutput.TempK, actualOutput.TempK, 0.01)

		mockLocation.AssertExpectations(t)
		mockWeather.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})
}
