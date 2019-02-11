# Accept the Go version for the image to be set as a build argument
ARG GO_VERSION=1.11.2

# First stage: build the executable
FROM golang:${GO_VERSION}-alpine AS builder

# Create appuser
RUN adduser -D -g '' appuser

# ca-certificates for calls to HTTPS endpoints
# git for fetching the dependencies
RUN apk add --no-cache \
	ca-certificates \
	git \
	gcc

RUN apk add musl-dev

# Get code
COPY . /src
WORKDIR /src

# Build the binary
RUN go build -o nut-upsd-exporter cmd/nut-upsd-exporter/main.go

# Second stage: build the container
FROM alpine

# Copy dependencies
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /src/nut-upsd-exporter /bin/nut-upsd-exporter

USER appuser
EXPOSE 8080
ENTRYPOINT ["/bin/nut-upsd-exporter"]

