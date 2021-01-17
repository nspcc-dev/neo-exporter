FROM golang:1.15-alpine as basebuilder
RUN apk add --update make bash

FROM basebuilder as builder
ARG VERSION=dev
WORKDIR /src
COPY . /src

RUN make bin

# Executable image
FROM scratch AS neofs-net-monitor

WORKDIR /

COPY --from=builder /src/bin/neofs-net-monitor /bin/neofs-net-monitor

CMD ["neofs-net-monitor"]