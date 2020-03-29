FROM golang:1.13.9 as builder
WORKDIR /app
COPY . .
WORKDIR /app/cmd/killboard
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine:latest AS release
WORKDIR /app

RUN apk --no-cache add tzdata
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/cmd/killboard/killboard .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"