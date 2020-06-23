package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"time"
)

type Card struct {
	PersonID       uint64    `json:"person_id"`
	AccountNumber  string    `json:"account_number"`
	CVC            string    `json:"cvc"`
	ID             string    `json:"id"`
	ExpirationDate time.Time `json:"expirationDate"`
}

type cardManagement struct {
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		return shim.Success(nil)
	},
	"delCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		return shim.Success(nil)
	},
	"getInfo": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		return shim.Success(nil)
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
