FABRIC_VERSION = 1.4.6
DOCKER = docker
HOME_DIR = /Users/dmitry/hlf-course
COMPOSE = docker-compose
GENESIS = genesis.block
CRYPTOGEN = cryptogen
CONFIGGEN = configtxgen


start-all: artifacts
	$(COMPOSE) -f docker-compose-cli.yaml up
stop-all:
	$(COMPOSE) -f docker-compose-cli.yaml down --volumes --remove-orphans
fabric-tools:
	$(DOCKER) run -it -d --rm  -v $(shell pwd):$(HOME_DIR)/ -w $(HOME_DIR) \
		--name fabric-tools -e FABRIC_ROOT=$(HOME_DIR)/ \
		-e FABRIC_CFG_PATH=$(HOME_DIR)	\
		hyperledger/fabric-tools:$(FABRIC_VERSION) bash
crypto: fabric-tools
	$(DOCKER) exec fabric-tools $(CRYPTOGEN) generate \
		--config=$(HOME_DIR)/crypto-config.yaml \
		--output=$(HOME_DIR)/crypto-config/.
artifacts: crypto
	$(DOCKER) exec fabric-tools $(CONFIGGEN) -profile TwoOrgsChannel \
		-channelID mychannel -outputCreateChannelTx=$(HOME_DIR)/channel-artifacts/channel.tx
	$(DOCKER) exec fabric-tools $(CONFIGGEN) -profile TwoOrgsOrdererGenesis \
		-outputBlock $(HOME_DIR)/channel-artifacts/$(GENESIS) -channelID systemchannel
clean-artifacts:
	$(RM) -rf channel-artifacts/*
	$(RM) -rf crypto-config/*
stop-tools:
	$(DOCKER) stop fabric-tools
clean-all: clean-artifacts stop-tools
