# Building with the official Go Alpine image
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main .


# Runtime image
FROM alpine:latest  

RUN apk add --no-cache libstdc++

WORKDIR /app
COPY --from=builder /app/main .
RUN mkdir -p /app/cache

EXPOSE 8080
VOLUME /app/cache
CMD ["./main", "-server"]