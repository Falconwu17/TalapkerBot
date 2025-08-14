FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY .. .

RUN go build -o api ./content-api/cmd/api

RUN adduser -D -u 10010 apiuser
USER apiuser

CMD ["./api"]
