################################
# STEP 1 build executable binary
################################
FROM golang:1.16 AS builder
# application source
WORKDIR /chicory
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -o bin/rscan cmd/rscan/main.go

############################
# STEP 2 build a small image
############################
FROM scratch

# include certs for HTTPS client
WORKDIR /etc/ssl/certs/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt .

# copy the executable from the builder
WORKDIR /
COPY --from=builder /chicory/bin/rscan .

# export port for HTTP endpoints
EXPOSE 8080

# Run the binary.
ENTRYPOINT ["./rscan"]
