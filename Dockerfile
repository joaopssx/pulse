FROM golang:1.22-alpine AS build

RUN apk add --no-cache build-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags='-s -w' \
    -tags 'sqlite_omit_load_extension' \
    -o /out/pulse ./cmd/pulse


FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget tzdata \
    && addgroup -S pulse \
    && adduser -S -u 10001 -G pulse pulse \
    && mkdir -p /app/data \
    && chown -R pulse:pulse /app

WORKDIR /app

COPY --from=build /out/pulse /app/pulse

USER pulse

EXPOSE 8080

ENTRYPOINT ["/app/pulse"]
