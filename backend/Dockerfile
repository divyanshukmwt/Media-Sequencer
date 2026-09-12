FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=direct
ENV GOSUMDB=off
ENV GOFLAGS=-mod=mod

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /media-sequencer .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /media-sequencer .

EXPOSE 8080

ENTRYPOINT ["./media-sequencer"]
