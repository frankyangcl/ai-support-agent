FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app

COPY --from=builder --chown=app:app /app/server ./server

RUN mkdir -p uploads && chown app:app uploads

USER app

EXPOSE 8080

CMD ["./server"]
