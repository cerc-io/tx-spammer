FROM golang:1.19-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers
# DEBUG
RUN apk add busybox-extras

# Get and build tx_spammer
ADD . /go/src/github.com/vulcanize/tx_spammer
WORKDIR /go/src/github.com/vulcanize/tx_spammer
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o tx_spammer .

# app container
FROM alpine

ARG USER="vdm"
ARG CONFIG_FILE="./environments/example.toml"

RUN adduser -Du 5000 $USER
WORKDIR /app
RUN chown $USER /app
USER $USER

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/tx_spammer/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/tx_spammer/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/tx_spammer/tx_spammer tx_spammer
COPY --from=builder /go/src/github.com/vulcanize/tx_spammer/environments environments

ENTRYPOINT ["/app/startup_script.sh"]
