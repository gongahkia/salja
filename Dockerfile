FROM golang:1.25-alpine AS builder
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /salja ./cmd/salja
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /salja-mcp ./cmd/salja-mcp
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
LABEL org.opencontainers.image.source="https://github.com/gongahkia/salja" \
      org.opencontainers.image.description="Universal calendar and task converter CLI"
COPY --from=builder /salja /usr/local/bin/salja
COPY --from=builder /salja-mcp /usr/local/bin/salja-mcp
ENTRYPOINT ["salja"]
