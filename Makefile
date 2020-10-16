## Build docker image
.PHONY: docker-build
docker-build:
	docker build -t vulcanize/tx_spammer -f Dockerfile .
