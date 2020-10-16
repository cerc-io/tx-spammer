#!/bin/sh

set -e
set +x
test $SPAMMER_COMMAND
set +e

echo "Running tx spammer"
./ipld-eth-indexer ${SPAMMER_COMMAND} --config=config.toml --log-file=${LOG_FILE} --log-level=${LOG_LEVEL}

if [ $? -eq 0 ]; then
    echo "tx spammer ran successfully"
else
    echo "tx spammer ran with error. Is the config file correct?"
    exit 1
fi