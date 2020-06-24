package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type BankAccount struct {
	PersonID      uint64  `json:person_id`
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
		response := stub.InvokeChaincode("persons-chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
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
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var account BankAccount

		AccountState, err := stub.GetState(params[0])

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account %s balance account due to %s", params[0], err))
		}

		if AccountState == nil {
			return shim.Error(fmt.Sprintf("Account with number %s doesnt exist", params[0]))
		}
		
		err = json.Unmarshal([]byte(AccountState), &account)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize bank account information error %s", err))
		}

		return shim.Success([]byte(fmt.Sprintf("%f", account.Balance)))
	},

	"delAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}
		
		AccountState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account %s state due to %s", params[0], err))
		}

		if AccountState == nil {
			return shim.Error(fmt.Sprintf("Account with number %s doesnt exist", params[0]))
		}

		err = stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete %s due to %s", params[0], err))
		}
	
		return shim.Success(nil)
	},

	"transfer": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 3 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}
		// param[0] - sensder, params[1] - reciever, params[2] - value

		var senderAccount BankAccount
		var recieverAccount BankAccount
		var transferValue float64
		
		SenderState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get sender state with account number %s due to %s", params[0], err))
		}

		if SenderState == nil {
			return shim.Error(fmt.Sprintf("Sender account with number %s doesnt exist", params[0]))
		}

		RecieverState, err := stub.GetState(params[1])

		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get reciever state with account number %s due to %s", params[0], err))
		}

		if RecieverState == nil {
			return shim.Error(fmt.Sprintf("Reciever account with number %s doesnt exist", params[0]))
		}

		err = json.Unmarshal([]byte(params[2]), &transferValue)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize transfer value information"))
		}

		if transferValue <= 0.0 {
			return shim.Error(fmt.Sprintf("Negative or zero value to transfer"))
		}

		err = json.Unmarshal([]byte(SenderState), &senderAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize senderAccount %s state due to %s", params[0], err))
		}

		err = json.Unmarshal([]byte(RecieverState), &recieverAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize recieverAccount %s state due to %s", params[1], err))
		}


		if senderAccount.Balance < transferValue {
			return shim.Error(fmt.Sprintf("Sender has not enough balance to trasfer"))
		}

		senderAccount.Balance = senderAccount.Balance - transferValue
		recieverAccount.Balance = recieverAccount.Balance + transferValue

		var newRecieverState []byte
		var newSenderState []byte

		newSenderState, err = json.Marshal(senderAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create senderAccount state due to %s", err))
		}

		newRecieverState, err = json.Marshal(recieverAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create recieverAccount state due to %s", err))
		}

		err = stub.PutState(senderAccount.AccountNumber, newSenderState)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to put new sender state due to %s", err))
		}

		err = stub.PutState(recieverAccount.AccountNumber, newRecieverState)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to put new reciever state due to %s", err))
		}
		
		return shim.Success(nil)
	},

	"changeBalance": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 2 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var account BankAccount
		var value float64

		accountState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account state with account number %s due to %s", params[0], err))
		}

		if accountState == nil {
			return shim.Error(fmt.Sprintf("Account with number %s doesnt exist", params[0]))
		}

		err = json.Unmarshal([]byte(params[1]), &value)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize transfer value information"))
		}

		err = json.Unmarshal([]byte(accountState), &account)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize account %s state due to %s", params[0], err))
		}

		if account.Balance + value < 0.0 {
			return shim.Error(fmt.Sprintf("Incorrect value"))
		}

		account.Balance = account.Balance + value
		var newAccountState []byte

		newAccountState, err = json.Marshal(account)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create account state due to %s", err))
		}

		err = stub.PutState(account.AccountNumber, newAccountState)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to put new account state due to %s", err))
		}	

		return shim.Success(nil)
	},
	
	"accountHistory": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		accountState, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account state with account number %s due to %s", params[0], err))
		}

		if accountState == nil {
			return shim.Error(fmt.Sprintf("Account with number %s doesnt exist", params[0]))
		}

		historyQII, err := stub.GetHistoryForKey(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to get account history due to %s", err))
		}

		var currentStepAccount BankAccount

		err = json.Unmarshal([]byte(accountState), &currentStepAccount)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to serialize account %s state due to %s", params[0], err))
		}

		var prevStepBalance float64
		var history = []string{""}

		prevStepBalance = currentStepAccount.Balance

		for historyQII.HasNext() {
			currentStep, err := historyQII.Next()
			if err != nil {
				return shim.Error(fmt.Sprintf("Unable to get history state"))
			}

			err = json.Unmarshal(currentStep.Value, &currentStepAccount)
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to serialize history state due to %s", err))
			}
			if currentStepAccount.Balance < prevStepBalance {
				history = append(history, fmt.Sprintf("+%f", (prevStepBalance - currentStepAccount.Balance)))
			} else if currentStepAccount.Balance > prevStepBalance {
				history = append(history, fmt.Sprintf("+%f", (currentStepAccount.Balance - prevStepBalance)))
			} else {
				history = append(history, fmt.Sprintf("0.0"))
			}
			prevStepBalance = currentStepAccount.Balance
		}
		
		return shim.Success([]byte(fmt.Sprintf("%s",history)))		
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
