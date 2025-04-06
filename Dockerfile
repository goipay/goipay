FROM golang:1.23-alpine AS builder

ARG TARGETARCH

RUN apk update && apk add --no-cache make wget

RUN wget -O /grpc-health-probe "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-$TARGETARCH"

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build


FROM alpine AS prod

WORKDIR /app

COPY --from=builder /app/bin/server .
COPY --from=builder /app/config.yml .
COPY --from=builder /grpc-health-probe /usr/local/bin/grpc-health-probe

RUN chmod +x /usr/local/bin/grpc-health-probe

EXPOSE 3000

ENTRYPOINT [ "./server" ]