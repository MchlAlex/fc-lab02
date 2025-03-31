package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/MchlAlex/fc-lab02/internal/entity"
)

// LocationFinder define a interface para buscar localização por CEP.
type LocationFinder interface {
	GetLocationByCEP(cep string) (string, error)
}

// ViaCEPService implementa LocationFinder usando a API ViaCEP.
type ViaCEPService struct {
	Client *http.Client
}

// NewViaCEPService cria uma nova instância de ViaCEPService.
func NewViaCEPService(client *http.Client) *ViaCEPService {
	if client == nil {
		client = http.DefaultClient
	}
	return &ViaCEPService{Client: client}
}

var (
	ErrInvalidCEPFormat = errors.New("invalid zipcode")
	ErrCEPNotFound      = errors.New("can not find zipcode")
)

// GetLocationByCEP busca a cidade correspondente a um CEP usando a API ViaCEP.
func (s *ViaCEPService) GetLocationByCEP(cep string) (string, error) {
	// 1. Validar formato do CEP (8 dígitos numéricos)
	match, _ := regexp.MatchString(`^\d{8}$`, cep)
	if !match {
		return "", ErrInvalidCEPFormat
	}

	// 2. Montar URL e fazer requisição
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create ViaCEP request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute ViaCEP request: %w", err)
	}
	defer resp.Body.Close()

	// 3. Ler e decodificar resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ViaCEP response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ViaCEP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var viaCEPResp entity.ViaCEPResponse
	err = json.Unmarshal(body, &viaCEPResp)
	if err != nil {
		// Verifica se o erro é devido a um CEP inválido retornado pela API
		if string(body) == "{\n  \"erro\": true\n}" || string(body) == "{\"erro\": true}" {
			return "", ErrCEPNotFound
		}
		return "", fmt.Errorf("failed to decode ViaCEP response: %w", err)
	}

	// 4. Verificar se o CEP foi encontrado pela API
	if viaCEPResp.Erro == "true" { // Compara com a string "true"
		return "", ErrCEPNotFound
	}

	if viaCEPResp.Localidade == "" {
		return "", fmt.Errorf("city name not found in ViaCEP response for CEP %s", cep)
	}

	return viaCEPResp.Localidade, nil
}
