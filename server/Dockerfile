FROM golang:latest AS builder
ADD . /app
WORKDIR /app
ENV GOOS=linux GOARCH=arm64
RUN go mod download
RUN go build -a -o /main .

FROM alpine
RUN apk add --no-cache ca-certificates openssl tzdata
COPY --from=builder /main ./
CMD ["./main"]
