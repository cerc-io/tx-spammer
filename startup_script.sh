#!/bin/sh

set +e
if [ -z "$SPAMMER_COMMAND" ]; then
  echo "SPAMMER_COMMAND env value is required!"
  exit 2
fi

if [ -z "$LOG_LEVEL" ]; then
  LOG_LEVEL="debug"
fi

if [ ! -z "$ACCOUNTS_CSV_URL" ] && [ ! -d "keys" ]; then
  mkdir keys
  wget -O accounts.csv $ACCOUNTS_CSV_URL
  i=0
  for line in `cat accounts.csv`; do
    echo $line | cut -d',' -f3 | sed 's/^0x//' > keys/$i.key
    i=$((i + 1))
  done
fi

if [ ! -d "accounts/addresses" ]; then
  mkdir -p "accounts/addresses"
fi

echo "Running tx spammer"
./tx-spammer ${SPAMMER_COMMAND} --config=config.toml --log-level=${LOG_LEVEL}

if [ $? -eq 0 ]; then
    echo "tx spammer ran successfully"
else
    echo "tx spammer ran with error. Is the config file correct?"
    exit 1
fi
