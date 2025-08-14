FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY .. .

RUN go build -o app ./bot/cmd/app

RUN adduser -D -u 10001 botuser
USER botuser

CMD ["./app"]
