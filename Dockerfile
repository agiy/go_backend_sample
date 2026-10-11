FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server .

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
RUN apk add --no-cache ca-certificates=20260909-r0 tzdata=2026e-r0 \
    && addgroup -S -g 10001 goravel && adduser -S -u 10001 -G goravel goravel
WORKDIR /app
COPY --from=builder /out/server ./server
COPY --from=builder /src/public ./public
COPY --from=builder /src/resources ./resources
RUN mkdir -p storage && chown -R goravel:goravel /app
USER 10001:10001
ENV APP_ENV=production APP_DEBUG=false APP_HOST=0.0.0.0 APP_PORT=3000
EXPOSE 3000
ENTRYPOINT ["./server"]
