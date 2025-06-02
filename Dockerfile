FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum .

RUN go mod tidy

COPY . .

RUN rm -f biosearch_engine
RUN go build -o biosearch_engine main.go
CMD ["./biosearch_engine"]
