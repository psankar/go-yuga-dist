# Build stage
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-yuga-dist .

# Final stage
FROM alpine:latest
WORKDIR /
COPY --from=builder /go-yuga-dist /go-yuga-dist
EXPOSE 8080
ENTRYPOINT ["/go-yuga-dist"]
