FROM golang:1.25-alpine AS builder
WORKDIR /src

# Copy modules manifests
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download

# Copy source
COPY . ./

RUN go build -o /out/nexus ./cmd/nexus

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/nexus /usr/local/bin/nexus
WORKDIR /data
ENTRYPOINT ["/usr/local/bin/nexus"]
