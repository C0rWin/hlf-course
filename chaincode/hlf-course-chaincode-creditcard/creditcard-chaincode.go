package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Person structure to capture single person
type CreditCard struct {
	CreditCardNumber  string `json:"credit_card_number"`
	ExpirationDte	  string `json:expiration_date`
	Cvc		  uint64 `json:"cvc"`
	PersonID  	  uint64 `json:"person_id"`
	AccountNumber  	  string `json:"account_number"`
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCreditCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of input parameters, expected 1, but got %d", len(params)))
		}

		var creditCard CreditCard
		if err := json.Unmarshal([]byte(params[0]), &creditCard); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
		}

		personID := fmt.Sprintf("%d", creditCard.PersonID)
		response := stub.InvokeChaincode("persons-chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s", personID))
		}

		response = stub.InvokeChaincode("bank-chaincode", [][]byte{[]byte("getBalance"), []byte(creditCard.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create credit for account with id %s", creditCard.AccountNumber))
		}

		creditCardState, err := stub.GetState(creditCard.CreditCardNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create credit card due to %s", err))
		}

		if creditCardState != nil {
			return shim.Error(fmt.Sprintf("credit card with number %s already exists", creditCard.CreditCardNumber))
		}

		err = stub.PutState(creditCard.CreditCardNumber, []byte(params[0]))
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to create credit card with number %s, due to %s", creditCard.CreditCardNumber, err))
		}

		return shim.Success(nil)
	},

	"getCreditCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of input parameters, expected 1, but got %d", len(params)))
		}

		creditCardState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read creditCard information, id %s, due to %s", params[0], err))
		}

		if creditCardState == nil {
			return shim.Error(fmt.Sprintf("Credit card with id %s doesn't exists", params[0]))
		}

		var creditCard CreditCard
		if err := json.Unmarshal([]byte(creditCardState), &creditCard); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
		}

		info := string(creditCardState)

		personID := fmt.Sprintf("%d", creditCard.PersonID)
		response := stub.InvokeChaincode("persons-chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to get person with id %s", personID))
		}

		info = info + string(response.GetPayload())

		response = stub.InvokeChaincode("bank-chaincode", [][]byte{[]byte("getBalance"), []byte(creditCard.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to get account with id %s", creditCard.AccountNumber))
		}

		info = info + string(response.GetPayload())

		return shim.Success([]byte(fmt.Sprintf(info)))				
	},

	"delCreditCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of input parameters, expected 1, but got %d", len(params)))
		}

		creditCardState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read creditCard information, id %s, due to %s", params[0], err))
		}

		if creditCardState == nil {
			return shim.Error(fmt.Sprintf("Credit card with id %s doesn't exists", params[0]))
		}

		err = stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete person information, id %s, due to %s", params[0], err))
		}
		return shim.Success(nil)
	},
}

type personManagement struct {
}

func (p *personManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Chaincode has been initialized")
	return shim.Success(nil)
}

func (p *personManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, params := stub.GetFunctionAndParameters()
	action, exists := actions[funcName]
	if !exists {
		return shim.Error("unknown operation")
	}

	return action(stub, params)
}

func main() {
	err := shim.Start(new(personManagement))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
