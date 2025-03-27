# Dockerfile

# ---- Build Stage ----
  FROM golang:1.24-alpine AS builder

  WORKDIR /app
  
  # Copie os arquivos de módulo Go e baixe as dependências
  COPY go.mod go.sum ./
  RUN go mod download
  
  # Copie o restante do código-fonte da aplicação
  COPY . .
  
  # Compile a aplicação Go
  # -o /app/server: Especifica o nome e local do executável de saída
  # -ldflags="-w -s": Reduz o tamanho do binário (opcional)
  # ./cmd/server: O pacote main a ser compilado
  RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server
  
  # ---- Runtime Stage ----
  # Use uma imagem base mínima para o runtime
  FROM alpine:latest
  
  # Defina o diretório de trabalho
  WORKDIR /app
  
  # Copie o executável compilado do estágio de build
  COPY --from=builder /app/server /app/server
  
  # Copie o arquivo .env (alternativa: usar variáveis de ambiente no Cloud Run)
  # Se for usar variáveis de ambiente no Cloud Run (recomendado), comente ou remova a linha abaixo
  COPY .env .
  
  # Exponha a porta que a aplicação vai escutar (deve corresponder à porta no código/config)
  EXPOSE 8080
  
  # Comando para executar a aplicação quando o contêiner iniciar
  # O Cloud Run injetará a variável de ambiente PORT, mas definimos um padrão aqui também.
  # O .env será lido automaticamente se presente no diretório de trabalho.
  CMD ["/app/server"]