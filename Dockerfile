FROM golang:1.21-alpine AS builder
WORKDIR /app
#
RUN apk --no-cache add ca-certificates
#
COPY . .
#
RUN go install -mod=mod github.com/ua-parser/uap-go/uaparser
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main
#
FROM scratch
#
COPY --from=builder /app/main /opt/go-pot/go-pot
COPY --from=builder /app/config.yml /opt/go-pot/config.yml
WORKDIR /opt/go-pot
CMD ["start", "--host", "0.0.0.0", "--config-file", "config.yml"]
ENTRYPOINT ["./go-pot"]
