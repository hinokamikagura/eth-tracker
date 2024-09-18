FROM golang:1.23.0-bookworm AS build

ENV CGO_ENABLED=1

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
WORKDIR /build

COPY . .
RUN go mod download
RUN go build -o celo-tracker-cache-bootstrap -ldflags="-X main.build=${BUILD} -s -w" cmd/bootstrap/main.go
RUN go build -o celo-tracker -ldflags="-X main.build=${BUILD} -s -w" cmd/service/*.go

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /service

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/* .

EXPOSE 5001

CMD ["./celo-tracker"]