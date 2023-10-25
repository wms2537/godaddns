FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETARCH
ADD . /app
WORKDIR /app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} GODEBUG=netdns=9
RUN go mod download
RUN go build -a -o /main .

FROM --platform=$TARGETPLATFORM alpine
RUN apk add --no-cache ca-certificates openssl tzdata
COPY --from=builder /main ./
CMD ["./main"]
