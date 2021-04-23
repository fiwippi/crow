FROM golang:1.16.3-alpine3.13 as builder

RUN apk --no-cache add tzdata

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o metrics

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/metrics /metrics

EXPOSE 8000
CMD ["/metrics"]