package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hlf-course/chaincode/helpers"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type BankAccount struct {
	PersonID      uint64  `json:"person_id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
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
	"delAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		accountId := params[0]

		accountState, err := stub.GetState(accountId)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account information due to %s", err))
		}
		if accountState == nil {
			return shim.Error(fmt.Sprintf("bank account with number %s doesn't exists", accountId))
		}

		if err := stub.DelState(accountId); err != nil {
			return shim.Error(fmt.Sprintf("failed to delete account id %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},
	"getBalance": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		var account BankAccount
		accountId := params[0]
		accountState, err := stub.GetState(accountId)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account information due to %s", err))
		}
		if accountState == nil {
			return shim.Error(fmt.Sprintf("bank account with number %s doesn't exists", accountId))
		}
		if err := json.Unmarshal(accountState, &account); err != nil {
			return shim.Error(fmt.Sprintf("failed to Unmarshal account due to %s", err))
		}

		return shim.Success(helpers.Float64ToByte(account.Balance))
	},
	"transfer": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
        if len(params) != 3 {
            return shim.Error(fmt.Sprintf("wrong number of parameters(use: from, to, amount)"))
        }

		var senderAccount BankAccount
		var receiverAccount BankAccount
        from := params[0]
        to := params[1]
        amount, err := strconv.ParseFloat(params[2], 64)
        if err != nil {
			return shim.Error(fmt.Sprintf("failed to convert amount string to number"))
		}

        senderState, err := stub.GetState(from)
        if err != nil {
            return shim.Error(fmt.Sprintf("failed to get sender account information due to %s", err))
        }
        if senderState == nil {
            return shim.Error(fmt.Sprintf("sender bank account with number %s doesn't exists", from))
        }
		if err := json.Unmarshal(senderState, &senderAccount); err != nil {
			return shim.Error(fmt.Sprintf("failed to read senderState %s, due to %s", senderState, err))
		}
        if senderAccount.Balance - amount < 0 {
			return shim.Error(fmt.Sprintf("sender doesn't have enough money"))
		}

		receiverState, err := stub.GetState(to)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get receiver account information due to %s", err))
		}
		if receiverState == nil {
			return shim.Error(fmt.Sprintf("receiver bank account with number %s doesn't exists", to))
		}
		if err := json.Unmarshal(receiverState, &receiverAccount); err != nil {
			return shim.Error(fmt.Sprintf("failed to read senderState %s, due to %s", senderState, err))
		}

		senderAccount.Balance -= amount
		receiverAccount.Balance += amount

		newSenderAccountByte, err := json.Marshal(senderAccount)
		if err != nil {
			shim.Error(fmt.Sprintf("failed to generate bytes of new sender, due to %s", err))
		}
		newReceiverAccountByte, err := json.Marshal(senderAccount)
		if err != nil {
			shim.Error(fmt.Sprintf("failed to generate bytes of new receiver, due to %s", err))
		}

		if err := stub.PutState(from, newSenderAccountByte); err != nil {
			shim.Error(fmt.Sprintf("failed to update balance of %s, due to %s", from, err))
		}
		if err := stub.PutState(to, newReceiverAccountByte); err != nil {
			shim.Error(fmt.Sprintf("failed to update balance of %s, due to %s", to, err))
		}

		return shim.Success(nil)
	},
	"getHistory": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		var history []string
		accountId := params[0]

		historyIterator, err := stub.GetHistoryForKey(accountId)
		if err != nil {
			shim.Error(fmt.Sprintf("failed to read history of %s, due to %s", accountId, err))
		}

		var currentHistoryAccount BankAccount
		var previousBalance float64
		isInitialRecord := true
		for historyIterator.HasNext() {
			response, err := historyIterator.Next()
			if err != nil {
				return shim.Error("failed get Next() for historyIterator")
			}

			err = json.Unmarshal(response.Value, &currentHistoryAccount)

			if isInitialRecord {
				history = append(history, fmt.Sprintf("%f", currentHistoryAccount.Balance))
			} else {
				diff := currentHistoryAccount.Balance - previousBalance
				if diff > 0 {
					history = append(history, fmt.Sprintf("+%f", diff))
				} else {
					history = append(history, fmt.Sprintf("%f", diff))
				}
			}

			previousBalance = currentHistoryAccount.Balance
			isInitialRecord = false
		}

		buf := &bytes.Buffer{}
		err = gob.NewEncoder(buf).Encode(history)
		if err != nil {
			return shim.Error("failed to encode history to []byte")
		}
		return shim.Success(buf.Bytes())
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
