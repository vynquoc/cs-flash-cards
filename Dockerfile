FROM golang:1.22.3-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o main ./cmd/api

RUN chmod +x main

EXPOSE 4000

CMD ["./main"]

