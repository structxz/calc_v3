FROM golang:1.24.3

WORKDIR /app

COPY . .

RUN go install github.com/grpc-ecosystem/grpc-health-probe@latest && \
    cp /go/bin/grpc-health-probe /usr/local/bin/
RUN go mod download
RUN go build -o orchestrator ./cmd/orchestrator
RUN go build -o agent ./cmd/agent

COPY .env .env

CMD ["sh", "-c", "echo 'Override CMD in docker-compose.yml'"]
