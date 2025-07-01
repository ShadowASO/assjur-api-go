FROM golang:1.24.4-bullseye

LABEL maintainer="Aldenor"

RUN apt-get update -qq && \
    apt-get install -y -qq \
      libtesseract-dev libleptonica-dev \
      tesseract-ocr-eng \
      tesseract-ocr-deu \
      tesseract-ocr-por \
      poppler-utils && \
    rm -rf /var/lib/apt/lists/*

ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/

# Diretório de trabalho dentro do container
WORKDIR /app

# Copiar arquivos de dependências do Go
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar o código no diretório atual para o diretório de trabalho dentro do container
COPY . .

# Criar diretório de logs com permissões
RUN useradd -m appuser
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs

# Compilar o binário da aplicação
RUN go build -v -o server ./cmd/main.go


# Expor a porta que a aplicação usa
EXPOSE 4001

# Comando para iniciar a aplicação
CMD ["./server"]