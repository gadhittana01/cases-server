# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go-modules first (needed for replace directive)
COPY go-modules ./go-modules

# Copy server go mod files
COPY server/go.mod server/go.sum ./

# Update replace directive to point to local go-modules
RUN sed -i 's|replace github.com/gadhittana01/cases-modules => ../go-modules|replace github.com/gadhittana01/cases-modules => ./go-modules|g' go.mod

# Download dependencies
RUN go mod download

# Copy server source code
COPY server/ .

# Update go.mod again after copying (in case it was overwritten)
RUN sed -i 's|replace github.com/gadhittana01/cases-modules => ../go-modules|replace github.com/gadhittana01/cases-modules => ./go-modules|g' go.mod

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server main.go wire_gen.go injector.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
COPY --from=builder /app/db/migration ./db/migration

EXPOSE 8000

CMD ["./server"]
