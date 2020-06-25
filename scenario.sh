#!/bin/bash

SLEEP_STEP=3
START_SLEEP=10
HLF_NETWORK_LOG_FILE=./fabricStart.log
CHANNEL_NAME=mychannel

function run {

    set -x;

    docker exec -it cli $@

    set +x;

    sleep $SLEEP_STEP;
}


function setGlobals {
  PEER=$1
  ORG=$2
  if [ $ORG -eq 1 ]; then
    CORE_PEER_LOCALMSPID="Org1MSP"
    CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
    CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    if [ $PEER -eq 0 ]; then
      CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    else
      CORE_PEER_ADDRESS=peer1.org1.example.com:8051
    fi
  elif [ $ORG -eq 2 ]; then
    CORE_PEER_LOCALMSPID="Org2MSP"
    CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
    CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    if [ $PEER -eq 0 ]; then
      CORE_PEER_ADDRESS=peer0.org2.example.com:9051
    else
      CORE_PEER_ADDRESS=peer1.org2.example.com:10051
    fi
  else
    echo "================== ORG Unknown =================="
  fi
}

function joinChannel {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG

  set -x
  docker exec -it -e CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID -e CORE_PEER_TLS_ROOTCERT_FILE=$CORE_PEER_TLS_ROOTCERT_FILE -e CORE_PEER_MSPCONFIGPATH=$CORE_PEER_MSPCONFIGPATH -e CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS \
        cli peer channel join -b ${CHANNEL_NAME}.block
  set +x

  sleep $SLEEP_STEP
}

function installChaincode {
  VERSION=0.1
  CONTRACT_NAME=$1
  CONTRACT_DIR=$2
  for org in 1 2; do
    for peer in 0 1; do
        setGlobals $peer $org
        set -x
        docker exec -it -e CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID -e CORE_PEER_TLS_ROOTCERT_FILE=$CORE_PEER_TLS_ROOTCERT_FILE -e CORE_PEER_MSPCONFIGPATH=$CORE_PEER_MSPCONFIGPATH -e CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS \
                cli peer chaincode install -n ${CONTRACT_NAME} -v ${VERSION} -p github.com/chaincode/${CONTRACT_DIR}/ 
        set +x
        echo "===================== Chaincode ${CONTRACT_NAME} is installed on peer$peer.org$org ===================== "
        echo
        sleep $SLEEP_STEP
    done
  done
}

function instantiateChaincode {
  CONTRACT_NAME=$1

  C_ARGUMENTS='{"Args":[]}'
  C_POLICY='OR("Org1MSP.peer","Org2MSP.peer")'
  VERSION=0.1
  for org in 1 2; do
    for peer in 0 1; do
        setGlobals $peer $org
        set -x
        docker exec -it -e CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID -e CORE_PEER_TLS_ROOTCERT_FILE=$CORE_PEER_TLS_ROOTCERT_FILE -e CORE_PEER_MSPCONFIGPATH=$CORE_PEER_MSPCONFIGPATH -e CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS \
                cli peer chaincode instantiate -o orderer.example.com:7050 -C $CHANNEL_NAME -n ${CONTRACT_NAME} -v ${VERSION} -c $C_ARGUMENTS -P $C_POLICY
        set +x
        echo "===================== Chaincode ${CONTRACT_NAME} is instantiated on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
        echo
        sleep $SLEEP_STEP
    done
  done
 
}

function joinChannels {
	for org in 1 2; do
	    for peer in 0 1; do
		joinChannel $peer $org
		echo "===================== peer${peer}.org${org} joined channel '$CHANNEL_NAME' ===================== "
		sleep $SLEEP_STEP
		echo
	    done
	done
}

# Запуск Фабрики
make start-all &>$HLF_NETWORK_LOG_FILE &

echo "######## Waiting $START_SLEEP sec till HL Fabric network is starting"

# Ожидание пока Фабрика запуститься и выполнит инициализацию
sleep $START_SLEEP;

# Создание канала
run 'peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/channel.tx'

# Подключение каждого пира каждой организации к данному каналу
# run 'peer channel join -b mychannel.block'
joinChannels

echo "===================== Chaincode installation ====================="

# Установка чейнкодов

installChaincode "personCC" "hlf-course-chaincode-person"
# run 'peer chaincode install -n personCC -p github.com/chaincode/hlf-course-chaincode-person/ -v 0.1'

installChaincode "bankCC" "hlf-course-chaincode-bank"
# run 'peer chaincode install -n bankCC -p github.com/chaincode/hlf-course-chaincode-bank/ -v 0.1'

installChaincode "cardCC" "hlf-course-chaincode-card"
# run 'peer chaincode install -n cardCC -p github.com/chaincode/hlf-course-chaincode-card/ -v 0.1'




echo "===================== Chaincode instantiation (only 1 will succeed due to permissive policy) ====================="

# Настройка чейнкодов

#C_ARGUMENTS='{"Args":[]}'
#C_POLICY='OR("Org1MSP.peer","Org2MSP.peer")'

instantiateChaincode "personCC"
# run "peer chaincode instantiate -n personCC -v 0.1 -C mychannel -c $C_ARGUMENTS -P $C_POLICY"

instantiateChaincode "bankCC"
# run "peer chaincode instantiate -n bankCC -v 0.1 -C mychannel -c $C_ARGUMENTS -P $C_POLICY"

instantiateChaincode "cardCC"
# run "peer chaincode instantiate -n cardCC -v 0.1 -C mychannel -c $C_ARGUMENTS -P $C_POLICY"



# Выполнение чейнкодов, в данном случае используются invoke, но можно и query (в случае, если не изменяется стейт)
echo "===================== Person Chaincode invocation ====================="

PCC_ADD='{"Args":["addPerson","{\"id\":2,\"first_name\":\"Nail\",\"second_name\":\"Iskhakov\",\"address\":\"Kumertau\",\"phone\":\"33242\"}"]}'
run "peer chaincode invoke -n personCC -C mychannel -c $PCC_ADD"

PCC_GET='{"Args":["getPerson","2"]}'
run "peer chaincode query -n personCC -C mychannel -c $PCC_GET" 



echo "===================== Bank Chaincode invocation ====================="

BCC_ADD='{"Args":["addAccount","{\"person_id\":2,\"account_number\":\"0\",\"balance\":100000.00}"]}'
run "peer chaincode invoke -n bankCC -C mychannel -c $BCC_ADD"

BCC_ADD2='{"Args":["addAccount","{\"person_id\":2,\"account_number\":\"123456\",\"balance\":0.00}"]}'
run "peer chaincode invoke -n bankCC -C mychannel -c $BCC_ADD2" 

BCC_GB='{"Args":["getBalance","123456"]}'
run "peer chaincode query -n bankCC -C mychannel -c $BCC_GB"

BCC_TR='{"Args":["transfer","{\"From\":\"0\",\"To\":\"123456\",\"Value\":1000}"]}'
run "peer chaincode invoke -n bankCC -C mychannel -c $BCC_TR"

run "peer chaincode invoke -n bankCC -C mychannel -c $BCC_TR"

BCC_GA='{"Args":["getAccount","123456"]}'
run "peer chaincode query -n bankCC -C mychannel -c $BCC_GA"

BCC_GH='{"Args":["getHistory","123456"]}'
run "peer chaincode query -n bankCC -C mychannel -c $BCC_GH"

BCC_GH0='{"Args":["getHistory","0"]}'
run "peer chaincode query -n bankCC -C mychannel -c $BCC_GH0"

echo "===================== Credit Card Chaincode invocation ====================="

CCC_ADD='{"Args":["addCard","{\"card_number\":\"1111222233334444\",\"expire_date\":\"06/20\",\"cvc\":111,\"person_id\":2,\"account_number\":\"123456\"}"]}'
run "peer chaincode invoke -n cardCC -C mychannel -c $CCC_ADD"

CCC_GCI='{"Args":["getCardInfo","1111222233334444"]}'
run "peer chaincode query -n cardCC -C mychannel -c $CCC_GCI"

CCC_DLC='{"Args":["delCard","1111222233334444"]}'
run "peer chaincode invoke -n cardCC -C mychannel -c $CCC_DLC"


# ./clean.sh
