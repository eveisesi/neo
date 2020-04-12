FROM golang:1.13.9 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o neo /app/cmd/neo

FROM alpine:latest AS release
WORKDIR /app

RUN apk --no-cache add tzdata
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/neo .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"