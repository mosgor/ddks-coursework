# ─── Stage 1: Builder ────────────────────────────────────────────────────────
FROM golang:1.25-bookworm AS builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/main cmd/app/main.go


# ─── Stage 2: Runtime с ONNX Runtime ─────────────────────────────────────────
FROM debian:bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    libc6 \
    libstdc++6 \
    libgcc-s1 \
    libgomp1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/lib/onnxruntime

COPY external/libonnxruntime.so /usr/lib/onnxruntime/libonnxruntime.so

RUN ln -sf /usr/lib/onnxruntime/libonnxruntime.so /usr/lib/onnxruntime/libonnxruntime.so.1

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main .
COPY migrations/ ./migrations/

RUN mkdir -p /app/models

ENV ONNXRUNTIME_SHARED_LIBRARY_PATH=/usr/lib/onnxruntime/libonnxruntime.so
ENV REC_MODEL_PATH=/app/models/user_tower.onnx
ENV EMBEDDING_DIM=384
ENV CGO_ENABLED=1

EXPOSE 8090

CMD ["./main", "-mode=server"]