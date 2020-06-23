package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Card struct {
	PersonID       string    `json:"person_id"`
	AccountNumber  string    `json:"account_number"`
	CVC            string    `json:"cvc"`
	ID             string    `json:"id"`
	ExpirationDate string    `json:"expirationDate"`
}

type cardManagement struct {
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var card Card
		cardBytes := []byte(params[0])
		if err := json.Unmarshal(cardBytes, &card); err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize card information, due to %s", err))
		}

		// Check if Person exists
		response := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(card.PersonID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to check if person with id %s exists", card.PersonID))
		}

		// Check if Person account exists
		response = stub.InvokeChaincode("bank_chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to check bank account with number %s", card.AccountNumber))
		}

		// Check if Card exists
		cardState, err := stub.GetState(card.ID)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create bank account, due to %s", err))
		}
		if cardState != nil {
			return shim.Error(fmt.Sprintf("card with number %s already exists", card.ID))
		}

		if err := stub.PutState(card.ID, cardBytes); err != nil {
			shim.Error(fmt.Sprintf("failed to save card with number %s, due to %s", card.ID, err))
		}

		return shim.Success(nil)
	},
	"delCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		cardId := params[0]

		cardState, err := stub.GetState(cardId)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get card information due to %s", err))
		}
		if cardState == nil {
			return shim.Error(fmt.Sprintf("card with number %s doesn't exists", cardId))
		}

		if err := stub.DelState(cardId); err != nil {
			return shim.Error(fmt.Sprintf("failed to delete card with number %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},
	"getInfo": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		res := ""

		var card Card
		cardNumber := params[0]
		cardState, err := stub.GetState(cardNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get card information, due to %s", err))
		}
		if cardState == nil {
			return shim.Error(fmt.Sprintf("Card with number %s doesn't exists", cardNumber))
		}
		if err := json.Unmarshal(cardState, &card); err != nil {
			return shim.Error(fmt.Sprintf("failed to Unmarshal card due to %s", err))
		}
		res += "Card: " + string(cardState)

		personResponse := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(card.PersonID)}, "mychannel")
		if personResponse.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to check if person with id %s exists", card.PersonID))
		}
		res += " Person: " + string(personResponse.GetPayload())

		accountResponse := stub.InvokeChaincode("bank_chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if accountResponse.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to check bank account with number %s", card.AccountNumber))
		}
		res += " Account: " + string(accountResponse.GetPayload())

		return shim.Success([]byte(fmt.Sprintf(res)))
	},
}

func (b cardManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Card Management chaincode is initialized")
	return shim.Success(nil)
}

func (b cardManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
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
