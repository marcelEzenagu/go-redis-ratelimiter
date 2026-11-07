FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./main.go


FROM alpine:3.20
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .

RUN adduser -D appuser
USER appuser

EXPOSE 8083


CMD ["./server"]
