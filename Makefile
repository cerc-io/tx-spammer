## Build docker image
.PHONY: docker-build
docker-build:
	docker build -t vulcanize/tx_spammer -f Dockerfile .

.PHONY: build
build:
	 GO111MODULE=on go build -o tx_spammer .

.PHONY: contract
contract:
	 cd sol && solc --abi --bin -o build --overwrite Test.sol
