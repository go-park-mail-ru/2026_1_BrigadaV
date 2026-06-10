FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

RUN mkdir -p uploads/photos
COPY assets/fonts ./assets/fonts
COPY assets/places ./assets/places

EXPOSE 8080

CMD ["./server"]
