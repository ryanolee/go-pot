FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

FROM builder as dev

RUN go install -mod=mod github.com/githubnemo/CompileDaemon
ENTRYPOINT /go/bin/CompileDaemon --build="go build -o /build/go-pot" --command="/build/go-pot start --host 0.0.0.0"

FROM builder as prod-build
ARG VERSION=Unknown
ARG COMMIT=Unknown
ARG DATE=Unknown

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main -ldflags="-s -w -X 'github.com/ryanolee/go-pot/cmd.version=${VERSION}' -X 'github.com/ryanolee/go-pot/cmd.commitHash=${COMMIT}' -X 'github.com/ryanolee/go-pot/cmd.buildDate=${DATE}'"

FROM scratch as prod

COPY --from=prod-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=prod-build /app/main /app/main
CMD ["start", "--host", "0.0.0.0"]
ENTRYPOINT ["/app/main"]