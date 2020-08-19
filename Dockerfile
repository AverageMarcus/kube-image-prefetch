FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.14 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/
ADD . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o kube-image-prefetch main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
WORKDIR /app/
COPY --from=builder /app/kube-image-prefetch /app/kube-image-prefetch
ENTRYPOINT ["/app/kube-image-prefetch"]
