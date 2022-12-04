FROM golang:latest AS builder
ADD . /app
WORKDIR /app
ENV GOOS=linux GOARCH=arm64
RUN go mod download
RUN CGO_ENABLED=0 go build -a -o /main .

FROM arm64v8/alpine
RUN apk add --no-cache ca-certificates openssl tzdata
COPY --from=builder /main ./
ENTRYPOINT ["./main"]
