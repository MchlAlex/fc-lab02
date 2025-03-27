# Desafio Go Weather API (fc-lab02)

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Required-blue?logo=docker)](https://www.docker.com/)

## Descrição

Este projeto implementa um sistema em Go que recebe um CEP (Código de Endereçamento Postal) brasileiro válido, identifica a cidade correspondente e retorna a temperatura atual do local em graus Celsius (°C), Fahrenheit (°F) e Kelvin (K).

A aplicação utiliza APIs externas para obter os dados:

1.  **ViaCEP (ou similar):** Para buscar a localização (cidade) a partir do CEP fornecido.
2.  **WeatherAPI (ou similar):** Para obter as informações meteorológicas (temperatura) da cidade encontrada.

O sistema foi desenvolvido para ser containerizado com Docker e publicado no Google Cloud Run.

## Tecnologias Utilizadas

*   **Go:** Linguagem de programação principal.
*   **Docker & Docker Compose:** Para containerização e orquestração local.
*   **ViaCEP API:** Para consulta de CEP.
*   **WeatherAPI:** Para consulta de clima.

## Pré-requisitos

*   Docker instalado: [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)
*   Docker Compose instalado: [https://docs.docker.com/compose/install/](https://docs.docker.com/compose/install/)
*   Uma chave de API válida do WeatherAPI ([https://www.weatherapi.com/](https://www.weatherapi.com/)).

## Como Executar Localmente (com Docker)

1.  **Clone o repositório:**
    ```bash
    git clone https://github.com/MchlAlex/fc-lab02.git
    cd fc-lab01
    ```

2.  **Configure as Variáveis de Ambiente:**
    *   Renomeie (ou copie) o arquivo `.env.example` para `.env`.
    *   Edite o arquivo `.env` e insira sua chave da WeatherAPI na variável `WEATHER_API_KEY`.
    ```dotenv
    # .env
    WEATHER_API_KEY=SUA_CHAVE_AQUI
    WEBSERVER_PORT=8080
    ```

3.  **Construa e suba os containers:**
    ```bash
    docker-compose up -d --build
    ```
    *   O serviço estará disponível na porta definida em `WEBSERVER_PORT` (padrão: 8080).

4.  **Faça uma requisição:**
    Use um cliente HTTP (como `curl`, Postman, ou seu navegador) para acessar o endpoint:
    ```bash
    curl http://localhost:8080/weather/{CEP_DESEJADO}
    ```
    *   Substitua `{CEP_DESEJADO}` por um CEP válido de 8 dígitos (sem hífen).
    *   Exemplo: `curl http://localhost:8080/weather/01001000`

## Endpoints da API

### `GET /weather/{cep}`

Busca a temperatura atual para a localização correspondente ao CEP fornecido.

*   **Parâmetros:**
    *   `cep` (na URL): CEP brasileiro de 8 dígitos (ex: `01001000`).

*   **Respostas:**
    *   **`200 OK`**: Sucesso. Retorna as temperaturas.
        ```json
        {
          "temp_C": 25.0,
          "temp_F": 77.0,
          "temp_K": 298.0
        }
        ```
    *   **`422 Unprocessable Entity`**: CEP inválido (formato incorreto).
        ```
        invalid zipcode
        ```
    *   **`404 Not Found`**: CEP não encontrado na API de consulta de CEP.
        ```
        can not find zipcode
        ```
    *   **`500 Internal Server Error`**: Erro interno no servidor (ex: falha ao contatar API externa, chave de API inválida, etc.). A mensagem de erro específica pode variar.

## Testes Automatizados

O projeto inclui testes automatizados localizados no diretório `/tests`. Para executá-los:

1.  **Certifique-se de ter o Go instalado** ou execute dentro do container Docker se preferir.
2.  **Navegue até o diretório raiz do projeto.**
3.  **Execute o comando de teste do Go:**
    ```bash
    go test ./...
    ```
    *   Este comando descobrirá e executará todos os testes dentro do projeto. Os testes cobrem as principais funcionalidades, incluindo:
        *   Validação de formato de CEP.
        *   Tratamento de CEP não encontrado.
        *   Consulta de clima e cálculo de temperaturas para um CEP válido.
        *   Respostas HTTP esperadas para cada cenário.

## Deploy no Google Cloud Run

Este projeto está preparado para deploy no Google Cloud Run. O `Dockerfile` define a imagem do container. Siga a documentação oficial do Google Cloud para realizar o deploy de um container: [https://cloud.google.com/run/docs/deploying](https://cloud.google.com/run/docs/deploying).

*   **URL da Aplicação no Cloud Run:** `https://fc-lab-01-43209898677.southamerica-east1.run.app`