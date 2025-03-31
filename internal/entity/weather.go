package entity

// ViaCEPResponse representa a resposta da API ViaCEP.
type ViaCEPResponse struct {
	Localidade string `json:"localidade"` // Nome da cidade
	Erro       string `json:"erro"`       // Indica se o CEP foi encontrado
}

// WeatherAPIResponse representa a parte relevante da resposta da API WeatherAPI.
type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"` // Temperatura em Celsius
	} `json:"current"`
}

// WeatherOutput representa a resposta final da nossa API.
type WeatherOutput struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"` // Temperatura em Celsius
	TempF float64 `json:"temp_F"` // Temperatura em Fahrenheit
	TempK float64 `json:"temp_K"` // Temperatura em Kelvin
}

// ErrorResponse representa uma resposta de erro padrão.
type ErrorResponse struct {
	Message string `json:"message"`
}
