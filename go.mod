module github.com/vulcanize/tx_spammer

go 1.13

require (
	github.com/ethereum/go-ethereum v1.9.10
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.0
	github.com/spf13/viper v1.7.1
)

replace github.com/ethereum/go-ethereum v1.9.10 => github.com/vulcanize/go-ethereum v1.9.10-optimism-0.0.2
