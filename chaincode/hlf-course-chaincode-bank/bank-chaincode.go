package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type BankAccount struct {
	PersonID      uint64  `json:"person_id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
}

type TransferData struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Value float64 `json:"value"`
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
			return shim.Error(fmt.Sprintf("failed to create bank account for person with id %s, due to %s", personID, response.Message))
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
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		state, err := stub.GetState(params[0])

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read account information, account number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("bank account with account number %s doesn't exists", params[0]))
		}

		var account BankAccount
		if err = json.Unmarshal([]byte(state), &account); err != nil {
			return shim.Error(fmt.Sprintf("failed to parse state data of bank account %s, data is corrupted, error %s", params[0], err))
		}

		return shim.Success([]byte(fmt.Sprintf("%f", account.Balance)))
	},
	"getAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read account information, account number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("bank account with account number %s doesn't exists", params[0]))
		}

		return shim.Success(state)
	},
	"delAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		err := stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete account information, id %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},

	"transfer": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		var transfer TransferData

		err := json.Unmarshal([]byte(params[0]), &transfer)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize transfer information error %s", err))
		}

		fromState, err := stub.GetState(transfer.From)

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account info %s", err))
		}

		if fromState == nil {
			return shim.Error(fmt.Sprintf("'from' account: %s doesn't exist", transfer.From))
		}

		var fromAccount BankAccount
		if err = json.Unmarshal([]byte(fromState), &fromAccount); err != nil {
			return shim.Error(fmt.Sprintf("failed to parse state data of bank 'from' account %s, data is corrupted, error %s", transfer.From, err))
		}

		toState, err := stub.GetState(transfer.To)

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account info %s", err))
		}

		if toState == nil {
			return shim.Error(fmt.Sprintf("'to' account: %s doesn't exist", transfer.To))
		}

		var toAccount BankAccount
		if err = json.Unmarshal([]byte(toState), &toAccount); err != nil {
			return shim.Error(fmt.Sprintf("failed to parse state data of bank 'to' account %s, data is corrupted, error %s", transfer.To, err))
		}

		// If fromAccount is zero number account, it can issue money - no limit on positiveness of balance
		if (fromAccount.Balance-transfer.Value) < 0 && fromAccount.AccountNumber != "0" {
			return shim.Error(fmt.Sprintf("Insufficient balance on 'from' account %s", transfer.From))
		}

		// Everything is OK
		fromAccount.Balance = fromAccount.Balance - transfer.Value
		toAccount.Balance = toAccount.Balance + transfer.Value

		fromString, err := json.Marshal(fromAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("Cannot serialize 'from' bank account struct, error %s", err))
		}

		toString, err := json.Marshal(toAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("Cannot serialize 'to' bank account struct, error %s", err))
		}

		err = stub.PutState(transfer.From, []byte(fromString))
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to store 'from' bank account with account number = %s, due to %s", transfer.From, err))
		}

		err = stub.PutState(transfer.To, []byte(toString))
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to store 'to' bank account with account number = %s, due to %s", transfer.To, err))
		}

		return shim.Success(nil)
	},

	"getHistory": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		historyIter, err := stub.GetHistoryForKey(params[0])

		if err != nil {
			return shim.Error(fmt.Sprintf("cannot retrieve history for key %s, due to %s", params[0], err))
		}

		result := make([]string, 0, 10)

		var historyEntry BankAccount

		prevBalance := 0.0

		for historyIter.HasNext() {
			modification, err := historyIter.Next()
			if err != nil {
				return shim.Error(fmt.Sprintf("cannot read record modification for key %s, due to %s", params[0], err))
			}
			err = json.Unmarshal(modification.Value, &historyEntry)
			if err != nil {
				return shim.Error(fmt.Sprintf("cannot unmarshal  history record %s, due tot error %s", modification.Value, err))
			}

			if math.Abs(historyEntry.Balance-prevBalance) > 0.0001 {
				result = append(result, fmt.Sprintf("%+f:%d", historyEntry.Balance-prevBalance, modification.Timestamp.GetSeconds()))
			}
			prevBalance = historyEntry.Balance
		}

		jsonString, _ := json.Marshal(result)
		return shim.Success([]byte(jsonString))
	},
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
