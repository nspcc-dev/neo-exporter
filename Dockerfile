FROM golang:1.25 as builder
ARG VERSION=dev
WORKDIR /src
COPY . /src

RUN make bin

# Executable image
FROM scratch AS neo-exporter

WORKDIR /

COPY --from=builder /src/bin/neo-exporter /bin/neo-exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["neo-exporter"]
