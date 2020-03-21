FABRIC_VERSION = 1.4.6
HOME_DIR = /home/hlf-course


fabric-tools:
	docker run -it -d --rm  -v $(shell pwd):$(HOME_DIR)/ -w $(HOME_DIR) \
		--name fabric-tools -e FABRIC_ROOT=/home/hlf-course/ \
		-e FABRIC_CFG_PATH=$(HOME_DIR)	\
		hyperledger/fabric-tools:$(FABRIC_VERSION) bash
crypto:
	docker exec fabric-tools cryptogen generate \
		--config=$(HOME_DIR)/crypto-config.yaml \
		--output=$(HOME_DIR)/crypto-config/.
artifacts:
	docker exec fabric-tools configtxgen -profile TwoOrgsChannel \
		-channelID mychannel -outputCreateChannelTx=$(HOME_DIR)/channel-artifacts/channel.tx
	docker exec fabric-tools configtxgen -profile TwoOrgsOrdererGenesis \
		-outputBlock /home/hlf-course/channel-artifacts/genesis.block -channelID systemchannel
clean-artifacts:
	rm -rf channel-artifacts/*
	rm -rf crypto-config/*
clean-all:
	rm -rf channel-artifacts/*
	rm -rf crypto-config/*
	docker-compose -f docker-compose-cli.yaml down --volumes --remove_orphans
	docker stop fabric-tools
