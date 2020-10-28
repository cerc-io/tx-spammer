# tx_spammer
Tools to enable the semi-reproducible growth of a large and complex chain over RPC, for testing and benchmarking purposes

Usage:

`./tx_spammer autoSend --config=./environments/gen.toml`

The `autoSend` command takes as input a .toml config of the below format, the fields can be overridden with the env variables in the comments.
It uses the provided key pairs and configuraiton parameters to generate and deploy a number of contracts with a simple interface for `Put`ing to a dynamic data structure.
It can then spam these contracts with `Put` operations at deterministically generated storage locations in order to grow the storage tries indefinitely and in a reproducible manner.
Additionally, it can be configured to send large numbers of eth value transfers to a set of manually designated or deterministically derived addresses,
in order to grow the state trie.

This process is semi-reproducible, in that given the same input it will generate and send the same set of contract deployment, eth transfer,
and contract calling transactions. The precise ordering of transactions is not guaranteed, and uncertainty is introduced by the
miner processing the transactions.

```toml
[eth]
    keyDirPath = "" # path to the directory with all of the key pairs to use - env: $ETH_KEY_DIR_PATH
    addrFilePath = "" # path to a file with a newline seperated list of addresses we want to send value transfers to - env: $ETH_ADDR_DIR_PATH
    httpPath = "" # http url for the node we wish to send all our transactions to - env: $ETH_HTTP_PATH
    chainID = 421 # chain id - env: $ETH_CHAIN_ID
    type = "L2" # tx type (EIP1559, Standard, or L2) - env: $ETH_TX_TYPE

[deployment]
    number = 1 # number of contracts we will deploy for each key at keyPath - env: $ETH_DEPLOYMENT_NUMBER
    hexData = "" # hex data for the contracts we will deploy - env: $ETH_DEPLOYMENT_HEX_DATA
    gasLimit = 0 # gasLimit to use for the deployment txs - env: $ETH_DEPLOYMENT_GAS_LIMIT
    gasPrice = "0" # gasPrice to use for the deployment txs - env: $ETH_DEPLOYMENT_GAS_PRICE

[optimism]
    l1Sender = "" # l1 sender address hex to use for all txs - env: $ETH_OPTIMISM_L1_SENDER
    l1RollupTxId = 0 # rollup tx id to use for all txs - env: $ETH_OPTIMISM_ROLLUP_TX_ID
    sigHashType = 0 # sig hash type to use for all txs - env: $ETH_OPTIMISM_SIG_HASH_TYPE
    queueOrigin = 0 # queue origin id to use for all txs - env: $ETH_OPTIMISM_QUEUE_ORIGIN

[contractSpammer]
    frequency = 30 # how often to send a transaction (in seconds) - env: $ETH_CALL_FREQ
    totalNumber = 10000 # total number of transactions to send (across all senders) - env: $ETH_CALL_TOTAL_NUMBER
    abiPath = "" # path to the abi file for the contract we are calling - env: $ETH_CALL_ABI_PATH
    # NOTE: we expect to be calling a method such as Put(address addr, uint256 val) where the first argument is an
    # integer than we can increment to store values at new locations in the contract trie (to grow it) and
    # the second argument is an integer value that we store at these positions
    methodName = "Put" # the method name we are calling - env: $ETH_CALL_METHOD_NAME
    storageValue =  1337 # the value we store at each position - env: $ETH_CALL_STORAGE_VALUE
    gasLimit = 0 # gasLimit to use for the eth call txs - env: $ETH_CALL_GAS_LIMIT
    gasPrice = "0" # gasPrice to use for the eth call txs - env: $ETH_CALL_GAS_PRICE

[sendSpammer]
    frequency = 30 # how often to send a transaction (in seconds) - env: $ETH_SEND_FREQ
    totalNumber = 10000 # total number of transactions to send (across all senders) - env: $ETH_SEND_TOTAL_NUMBER
    amount = "1" # amount of wei (1x10^-18 ETH) to send in each tx (be mindful of the genesis allocations) - env: $ETH_SEND_AMOUNT
    gasLimit = 0 # gasLimit to use for the eth transfer txs - env: $ETH_SEND_GAS_LIMIT
    gasPrice = "0" # gasPrice to use for the eth transfer txs - env: $ETH_SEND_GAS_PRICE
```

TODO: better documentation and document the other commands