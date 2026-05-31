FROM golang:1.26.0-alpine AS builder

LABEL maintainer="Aldenor"

# Instalar dependências necessárias para build (se precisar)
RUN apk add --no-cache \
    build-base \
    poppler-utils

# Diretório de trabalho dentro do container
WORKDIR /app

# Copiar arquivos de dependências do Go
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar o código no diretório atual para o diretório de trabalho dentro do container
COPY . .


# Compilar o binário da aplicação
RUN go build -v -o server ./cmd/server.go

#------------------------------------------------------------
#    CONCLUÍDA A COMPILAÇÃO - SEGUE A CÓPIA PARA O ALPINE
#------------------------------------------------------------    

FROM alpine:latest


WORKDIR /app

# Instalar poppler-utils na imagem final (necessário em runtime)
#RUN apk add --no-cache poppler-utils
RUN apk add --no-cache \
    poppler-utils \
    tzdata

ENV TZ=America/Fortaleza

RUN mkdir -p /app/logs

COPY --from=builder /app/server .
COPY --from=builder /app/.env .


# Expor a porta que a aplicação usa
EXPOSE 4001

# Comando para iniciar a aplicação
CMD ["./server"]