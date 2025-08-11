# Dockerfile
FROM golang:1.24-alpine

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bot


RUN adduser -D -u 10001 botuser
USER botuser
CMD ["/app/bot"]
