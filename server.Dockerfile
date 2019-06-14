FROM golang:1.12-alpine3.9 as builder

RUN apk add --no-cache git ca-certificates

WORKDIR /gomod

COPY . .

RUN CGO_ENABLED=0 go build -mod vendor -o go-weather-app ./server

FROM alpine:3.9 as development

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /gomod/go-weather-app /go-weather-app

ENTRYPOINT [ "/go-weather-app" ]

FROM scratch as production

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /gomod/go-weather-app /go-weather-app

ENTRYPOINT [ "/go-weather-app" ]
