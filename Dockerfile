FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY . .
RUN go mod tidy

# Compile
RUN go build -o devsecops-api ./api

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/devsecops-api .

EXPOSE 8080
CMD ["./devsecops-api"]
