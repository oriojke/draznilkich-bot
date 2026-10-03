FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/bot /bot
USER 65532:65532
ENTRYPOINT ["/bot"]
