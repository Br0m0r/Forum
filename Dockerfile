# Stage 1: Build with GCC and Go
FROM golang:1.23.0-alpine AS builder

# Install build tools for CGO and SQLite
RUN apk add --no-cache build-base sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .



# Build with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o forum

# Stage 2: Minimal Alpine with required libs
FROM alpine:3.19

# Install runtime libs needed for CGO SQLite
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/forum .

# Copy the static folder (which includes templates)
COPY --from=builder /app/static /app/static
COPY --from=builder /app/db /app/db

EXPOSE 8080

CMD ["./forum"]
