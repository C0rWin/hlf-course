package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Bad practice, this struct is originally located in hlf-course-chaincode-bank
type BankAccount struct {
	PersonID      uint64  `json:"person_id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
}

type DetailedInfo struct {
	Person     string `json:"person"`
	CreditCard string `json:"credit_card"`
	Account    string `json:"account"`
}

type CreditCard struct {
	CardNumber    string `json:"card_number"`
	ExpireDate    string `json:"expire_date"`
	CVC           uint   `json:"cvc"`
	PersonID      uint64 `json:"person_id"`
	AccountNumber string `json:"account_number"`
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of input parameters, expected 1, but got %d", len(params)))
		}

		var card CreditCard
		err := json.Unmarshal([]byte(params[0]), &card)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize credit card information error %s", err))
		}

		// Checking existance of person
		personID := fmt.Sprintf("%d", card.PersonID)
		response := stub.InvokeChaincode("personCC", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s, due to %s", personID, response.Message))
		}

		// Checking existance of bank account
		response = stub.InvokeChaincode("bankCC", [][]byte{[]byte("getAccount"), []byte(card.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit cart and link it to the bank account %s, due to %s", card.AccountNumber, response.Message))
		}

		var account BankAccount
		err = json.Unmarshal(response.GetPayload(), &account)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to deserialize payload recieved from bankCC chaincode: recieved data - %s, error - %s", response.GetPayload(), err))
		}

		// Checking wheather specified client is owner of the specified bank account
		if account.PersonID != card.PersonID {
			return shim.Error(fmt.Sprintf("unable to create credit card due to specified person - %s is not an owner of the bank account - %s", personID, account.AccountNumber))
		}

		// Checking existance of the card number and that we do not create a duplicate
		state, err := stub.GetState(card.CardNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read state for key %s, error %s", card.CardNumber, err))
		}

		if state != nil {
			return shim.Error(fmt.Sprintf("card with number %s already exist", card.CardNumber))
		}

		// err = stub.PutState(card.CardNumber, []byte(params[0]))
		jsonString, _ := json.Marshal(card)
		err = stub.PutState(card.CardNumber, jsonString)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to store credit card with number = %s, due to %s", card.CardNumber, err))
		}

		return shim.Success([]byte(card.CardNumber))
	},

	"delCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {

		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		err := stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete credit card, number %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},

	"getCardInfo": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read state for key %s, error %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("card with number %s does not exist", params[0]))
		}

		var card CreditCard
		err = json.Unmarshal(state, &card)
		if err != nil {
			return shim.Error(fmt.Sprintf("unable to deserialize state data of card %s, error %s", state, err))
		}

		// Checking existance of person
		personID := fmt.Sprintf("%d", card.PersonID)
		responsePerson := stub.InvokeChaincode("personCC", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if responsePerson.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s, due to %s", personID, responsePerson.Message))
		}

		// Checking existance of bank account
		responseAccount := stub.InvokeChaincode("bankCC", [][]byte{[]byte("getAccount"), []byte(card.AccountNumber)}, "mychannel")
		if responseAccount.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit cart and link it to the bank account %s, due to %s", card.AccountNumber, responseAccount.Message))
		}

		detailedInfo := DetailedInfo{
			Account:    string(responseAccount.GetPayload()),
			Person:     string(responsePerson.GetPayload()),
			CreditCard: string(state),
		}

		jsonString, _ := json.Marshal(detailedInfo)

		return shim.Success(jsonString)
	},
}

type creditCardManagement struct {
}

func (b creditCardManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Credit Card Management chaincode is initialized")
	return shim.Success(nil)
}

func (b creditCardManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, params := stub.GetFunctionAndParameters()
	action, exists := actions[funcName]
	if !exists {
		return shim.Error("unknown operation")
	}

	return action(stub, params)
}

func main() {
	err := shim.Start(new(creditCardManagement))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
