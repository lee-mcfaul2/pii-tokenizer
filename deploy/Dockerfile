# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags='-s -w' -o /out/tokenizer ./cmd/tokenizer && \
    go build -trimpath -ldflags='-s -w' -o /out/rotate-kmaster ./tools/rotate-kmaster

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/tokenizer /usr/local/bin/tokenizer
COPY --from=builder /out/rotate-kmaster /usr/local/bin/rotate-kmaster
EXPOSE 8443
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/tokenizer"]
