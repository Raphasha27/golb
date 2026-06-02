FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o golb ./cmd/golb

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/golb /usr/local/bin/golb
COPY config.yaml /etc/golb/config.yaml
ENTRYPOINT ["golb", "--config", "/etc/golb/config.yaml"]
