FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.26rc2 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/
ADD . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o tailon main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.23.3

RUN mkdir -p /app/data/ && adduser -D -u 1000 tailon -h /app/data
VOLUME /app/data
USER nonroot

WORKDIR /app/
COPY --from=builder /app/tailon /app/tailon
ENTRYPOINT ["/app/tailon"]