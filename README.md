# This is an example of creating Hyperledger Fabric network


## Run fabric tools instead of download binaries:

```
   docker run -it -d --rm  -v $(pwd):/home/hlf-course/ -w /home/hlf-course --name 
   fabric-tools -e FABRIC_ROOT=/home/hlf-course/ -e FABRIC_CFG_PATH=/home/hlf-course/  
   hyperledger/fabric-tools bash
```

## Generate crypto materials:

```
   docker exec fabric-tools cryptogen generate 
   --config=/home/hlf-course/crypto-config.yaml --output=/home/hlf-course/crypto-config/.
```

## Generate system channel genesis.block:

```
   docker exec fabric-tools configtxgen -profile TwoOrgsOrdererGenesis -outputBlock 
   /home/hlf-course/channel-artifacts/genesis.block -channelID systemchannel
```

## Generate create channel transaction:

```
  docker exec fabric-tools configtxgen -profile TwoOrgsChannel -channelID mychannel 
  -outputCreateChannelTx=/home/hlf-course/channel-artifacts/channel.tx
```


## Span Fabric network

```
  docker-compose -f docker-compose-cli.yaml up
```

## Login into cli container

```
  docker attach cli
```

## Create channel 

```
  peer channel create -o orderer.example.com:7050 -c mychannel -f channel-artifacts/channel.tx
```

## Join channel

```
  peer channel join -b mychannel.block
```
