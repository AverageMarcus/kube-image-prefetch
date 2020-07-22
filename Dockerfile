FROM golang:1.14-alpine AS builder
WORKDIR /app/
ADD . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o kube-image-prefetch main.go

FROM scratch
WORKDIR /app/
COPY --from=builder /app/kube-image-prefetch /app/kube-image-prefetch
ENTRYPOINT ["/app/kube-image-prefetch"]
