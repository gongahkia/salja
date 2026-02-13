FROM golang:1.25 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /salja ./cmd/salja

FROM scratch
COPY --from=builder /salja /salja
ENTRYPOINT ["/salja"]
