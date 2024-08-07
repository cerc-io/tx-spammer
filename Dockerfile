FROM golang:1.21-alpine as builder

RUN apk --update --no-cache add gcc libc-dev
# DEBUG
RUN apk add busybox-extras

# Get and build tx-spammer
WORKDIR /go/src/github.com/cerc-io/tx-spammer

ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o tx-spammer .

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
COPY --chown=5000:5000 --from=builder /go/src/github.com/cerc-io/tx-spammer/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/cerc-io/tx-spammer/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/cerc-io/tx-spammer/tx-spammer tx-spammer
COPY --from=builder /go/src/github.com/cerc-io/tx-spammer/environments environments
COPY --from=builder /go/src/github.com/cerc-io/tx-spammer/sol sol

ENTRYPOINT ["/app/startup_script.sh"]
