# This is an example of creating Hyperledger Fabric network


## Run fabric tools instead of download binaries:

```
   docker run -it -d --rm  -v $(pwd):/home/hlf-course/ -w /home/hlf-course --name fabric-tools -e FABRIC_ROOT=/home/hlf-course/ -e FABRIC_CFG_PATH=/home/hlf-course/ hyperledger/fabric-tools bash
```

## Generate crypto materials:

```
   docker exec fabric-tools cryptogen generate --config=/home/hlf-course/crypto-config.yaml --output=/home/hlf-course/crypto-config/.
```

## Generate system channel genesis.block:

```
   docker exec fabric-tools configtxgen -profile TwoOrgsOrdererGenesis -outputBlock /home/hlf-course/channel-artifacts/genesis.block -channelID systemchannel
```

>CRIT 008 Error on outputBlock: Error writing genesis block: open ./channel-artifacts/genesis.block: no such file or directory
><br>
>Solution : Create channel-artifacts folder before you run the command

## Generate create channel transaction:

```
  docker exec fabric-tools configtxgen -profile TwoOrgsChannel -channelID mychannel -outputCreateChannelTx=/home/hlf-course/channel-artifacts/channel.tx
```


## Span Fabric network

```
  docker-compose -f docker-compose-cli.yaml up -d
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

## Install chaincode
```
  peer chaincode install -p github.com/chaincode/hlf-course-chaincode-person -n persons_chaincode -v 1.0
  OR
  peer chaincode install -p github.com/chaincode/hlf-course-chaincode-bank -n bank_chaincode -v 1.0
```

## Instantiate chaincode
```
  peer chaincode instantiate -n persons_chaincode -v 1.0 -o orderer.example.com:7050 -C mychannel -c '{"Args":[""]}' -P "OR ('Org1MSP.peer', 'Org2MSP.peer')"
  OR
  peer chaincode instantiate -n bank_chaincode -v 1.0 -o orderer.example.com:7050 -C mychannel -c '{"Args":[""]}' -P "OR ('Org1MSP.peer', 'Org2MSP.peer')"
```

## Invoke chaincode
```
  peer chaincode invoke -n persons_chaincode -C mychannel -c '{"Args":["addPerson","{\"id\":1,\"first_name\":\"Dmitry\",\"second_name\":\"Kudryavtsev\",\"address\":\"Home\",\"phone\":\"88005553535\"}"]}'
  peer chaincode invoke -n persons_chaincode -C mychannel -c '{"Args":["getPerson", "1"]}'
```
