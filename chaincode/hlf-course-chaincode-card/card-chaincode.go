package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Card struct {
	CardNumber    string `json:"card_number"`
	ValidThrough  uint64 `json:"valid_through"`
	Cvc           string `json:"cvc"`
	PersonID      uint64 `json:"person_id"`
	AccountNumber string `json:"account_number"`
}

type cardManagement struct {
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var card Card
		err := json.Unmarshal([]byte(params[0]), &card)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize card information error %s", err))
		}

		personID := strconv.FormatUint(card.PersonID, 10)
		response := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s, due to %s", personID, err))
		}

		response = stub.InvokeChaincode("bank-chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit card for account number %s, due to %s", card.AccountNumber, err))
		}

		cardState, err := stub.GetState(card.CardNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create card due to %s", err))
		}

		if cardState != nil {
			return shim.Error(fmt.Sprintf("card with number %s already exists", card.CardNumber))
		}

		if err := stub.PutState(card.CardNumber, []byte(params[0])); err != nil {
			shim.Error(fmt.Sprintf("failed to save card with number %s, due to %s", card.CardNumber, err))
		}

		return shim.Success(nil)
	},

	"getAllInfo": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		cardState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get card info due to %s", err))
		}

		if cardState == nil {
			return shim.Error(fmt.Sprintf("card with number %s doesn't exist", params[0]))
		}

		var card Card
		err = json.Unmarshal([]byte(cardState), &card)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize card information error %s", err))
		}

		personID := strconv.FormatUint(card.PersonID, 10)
		personResponse := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if personResponse.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to get information about person with id %s, due to %s", personID, err))
		}

		accountResponse := stub.InvokeChaincode("bank-chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if accountResponse.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to get info for account number %s, due to %s", card.AccountNumber, err))
		}

		return shim.Success([]byte(fmt.Sprintf(string(personResponse.GetPayload()) + string(cardState) + " Balance: " + string(accountResponse.GetPayload()))))
	},

	"delCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		err := stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete card, number %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},
}

func (c *cardManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Card Management chaincode is initialized")
	return shim.Success(nil)
}

func (c *cardManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, params := stub.GetFunctionAndParameters()
	action, exists := actions[funcName]
	if !exists {
		return shim.Error("unknown operation")
	}
	return action(stub, params)
}

func main() {
	err := shim.Start(new(cardManagement))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

