package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type BankAccount struct {
	PersonID      string  `json:person_id`
	AccountNumber string  `json:account_number`
	Balance       float64 `json:balance`
}

type bankManagement struct {
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var account BankAccount
		err := json.Unmarshal([]byte(params[0]), &account)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize bank account information error %s", err))
		}

		// Need to check whenever account.PersonID is exists
		personID := fmt.Sprintf("%d", account.PersonID)
		response := stub.InvokeChaincode("personCC", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
		if response.Status == shim.ERROR {
			return shim.Error(fmt.Sprintf("failed to create bank account for person with id %s, due to %s", personID, err))
		}

		accountState, err := stub.GetState(account.AccountNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create bank account due to %s", err))
		}

		if accountState != nil {
			return shim.Error(fmt.Sprintf("bank account with number %s already exists", account.AccountNumber))
		}

		if err := stub.PutState(account.AccountNumber, []byte(params[0])); err != nil {
			shim.Error(fmt.Sprintf("failed to save bank account with number %s, due to %s", account.AccountNumber, err))
		}

		return shim.Success(nil)
	},

	"getBalance": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read bank account information with number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("bank account with number %s doesn't exists", params[0]))
		}
		var account BankAccount
		if err := json.Unmarshal([]byte(state), &account); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", state, err))
		}

		return shim.Success(account.Balance)
	},

	"delAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read bank account information with number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("bank account with number %s doesn't exists", params[0]))
		}

		err := stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete bank account information with number %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},

	"transfer": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		var senderAccount BankAccount
		var recipientAccount BankAccount
		var amountToTransfer float64

		if len(params) != 3 {
			return shim.Error(fmt.Sprintf("wrong number of parameters. Usage: senderAccountNumber recipientAccountNumber amountToTransfer"))
		}

		if err := json.Unmarshal([]byte(params[2]), &amountToTransfer); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[2], err))
		}
		
		if params[0] != "0" {
			senderState, err := stub.GetState(params[0])
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to read sender bank account information with number %s, due to %s", params[0], err))
			}

			if senderState == nil {
				return shim.Error(fmt.Sprintf("sender bank account with number %s doesn't exists", params[0]))
			}
			if err := json.Unmarshal([]byte(senderState), &account); err != nil {
				return shim.Error(fmt.Sprintf("failed to read input %s, error %s", senderState, err))
			}

			if senderAccount.Balance-amountToTransfer < 0 {
				return shim.Error(fmt.Sprintf("insufficient funds from the sender bank account with number %s. Sender has: %f , but want to transfer %f .", senderAccount.AccountNumber, senderAccount.Balance, amountToTransfer))
			}
		}

		recipientState, err := stub.GetState(params[1])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read recipient bank account information with number %s, due to %s", params[1], err))
		}

		if recipientState == nil {
			return shim.Error(fmt.Sprintf("recipient bank account with number %s doesn't exists", params[1]))
		}

		if err := json.Unmarshal([]byte(recipientState), &recipientAccount); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", recipientState, err))
		}

		if recipientAccount.Balance+amountToTransfer < recipientAccount.Balance {
			return shim.Error(fmt.Sprintf("restricted operation with negative amount to transfer"))
		}

		senderAccount.Balance = senderAccount.Balance - amountToTransfer
		recipientAccount.Balance = recipientAccount.Balance + amountToTransfer
		senderBytes, err := json.Marshal(senderAccount)
		recipientBytes, err := json.Marshal(recipientAccount)

		if err := stub.PutState(senderAccount.AccountNumber, senderBytes); err != nil {
			shim.Error(fmt.Sprintf("failed to save bank account with number %s, due to %s", senderAccount.AccountNumber, err))
		}
		if err := stub.PutState(recipientAccount.AccountNumber, recipientBytes); err != nil {
			shim.Error(fmt.Sprintf("failed to save bank account with number %s, due to %s", recipientAccount.AccountNumber, err))
		}

		return shim.Success(nil)

	}
}

func (b bankManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Bank Management chaincode is initialized")
	return shim.Success(nil)
}

func (b bankManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, params := stub.GetFunctionAndParameters()
	action, exists := actions[funcName]
	if !exists {
		return shim.Error("unknown operation")
	}

	return action(stub, params)
}

func main() {
	err := shim.Start(new(bankManagement))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
