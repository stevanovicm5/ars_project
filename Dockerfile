FROM golang:1.25.1-alpine AS build

RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=build /app/app .
COPY --from=build /app/docs ./docs

EXPOSE 8080

CMD ["./app"]