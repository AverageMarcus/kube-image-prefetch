FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx:golang AS xgo
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.14-alpine AS builder
COPY --from=xgo / /
WORKDIR /app/
ADD . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o kube-image-prefetch main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
WORKDIR /app/
COPY --from=builder /app/kube-image-prefetch /app/kube-image-prefetch
ENTRYPOINT ["/app/kube-image-prefetch"]
