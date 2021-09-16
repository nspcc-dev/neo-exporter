FROM golang:1.16 as builder
ARG VERSION=dev
WORKDIR /src
COPY . /src

RUN make bin

# Executable image
FROM scratch AS neofs-net-monitor

WORKDIR /

COPY --from=builder /src/bin/neofs-net-monitor /bin/neofs-net-monitor
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["neofs-net-monitor"]
