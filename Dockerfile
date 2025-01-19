FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/registry .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/registry /app/registry

EXPOSE 5050
CMD [ "/app/registry" ]