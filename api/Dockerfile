FROM golang:1-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

CMD [ "air", "--build.cmd", "go run ./cmd/api/main.go" ]