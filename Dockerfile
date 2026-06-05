FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY . .
RUN go mod tidy

# Compile
RUN go build -o scalabit-challenge-api ./api

FROM alpine:latest
WORKDIR /app

RUN addgroup -S appgroup && adduser -S nonroot -G appgroup

COPY --from=builder /app/scalabit-challenge-api .
COPY --from=builder /app/static ./static

USER nonroot

EXPOSE 8080
CMD ["./scalabit-challenge-api"]
