FROM golang:1.21-alpine AS builder
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

FROM builder as dev

RUN go install -mod=mod github.com/githubnemo/CompileDaemon
ENTRYPOINT /go/bin/CompileDaemon --build="go build -o /build/main" --command="/build/main start --port 80"

FROM builder as prod-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main

FROM scratch as prod

COPY --from=prod-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=prod-build /app/main /app/main

EXPOSE 80 7947
ENTRYPOINT ["/app/main", "start", "--port", "80"]