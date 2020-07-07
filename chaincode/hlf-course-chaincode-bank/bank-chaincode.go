package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type BankAccount struct {
	PersonID      uint64  `json:"person_id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
}

type TransferParams struct {
	AccountFrom string  `json:"account_from"`
	AccountTo   string  `json:"account_to"`
	Value       float64 `json:"value"`
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
			return shim.Error(fmt.Sprintf("failed to deserialize bank account information error %s", err))
		}

		// Need to check whenever account.PersonID is exists
		personID := strconv.FormatUint(account.PersonID, 10)
		response := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(personID)}, "mychannel")
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
			return shim.Error(fmt.Sprintf("failed to save bank account with number %s, due to %s", account.AccountNumber, err))
		}

		return shim.Success(nil)
	},

	"getBalance": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}
    
    state, err := stub.GetState(params[0])
    if err != nil {
      return shim.Error(fmt.Sprintf("failed to read account %s information, due to %s", params[0], err))
    }
    
    if state == nil {
      return shim.Error(fmt.Sprintf("account with account number %s doesn't exists", params[0]))
    }
    
    var account BankAccount
    
    err = json.Unmarshal([]byte(state), &account)
	  if err != nil {
		  return shim.Error(fmt.Sprintf("failed to deserialize bank account information error %s", err))
	  }	

		return shim.Success([]byte(strconv.FormatFloat(account.Balance, 'f', 2, 64)))
	},

	"delAccount": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

    state, err := stub.GetState(params[0])
    if err != nil {
      return shim.Error(fmt.Sprintf("failed to read account %s information, due to %s", params[0], err))
    }
    
    if state == nil {
      return shim.Error(fmt.Sprintf("account with account number %s doesn't exists", params[0]))
    }

		err = stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete account information, id %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},

	"transfer": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}
		var tParams TransferParams
    var account BankAccount
    b := true

		err := json.Unmarshal([]byte(params[0]), &tParams)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize transfer parameters information error %s", err))
		}

    if tParams.AccountFrom == "0"{
      b = false
    }

		if b {
			state, err := stub.GetState(tParams.AccountFrom)
      if err != nil {
        return shim.Error(fmt.Sprintf("failed to read accountFrom %s information, due to %s", tParams.AccountFrom, err))
      }
    
      if state == nil {
        return shim.Error(fmt.Sprintf("accountFrom with account number %s doesn't exists", tParams.AccountFrom))
      } 
      
      err = json.Unmarshal([]byte(state), &account)
	    if err != nil {
		    return shim.Error(fmt.Sprintf("failed to deserialize bank accountFrom information error %s", err))
	    }

      if account.Balance < tParams.Value {
			  return shim.Error(fmt.Sprintf("not enough money at bank account %s to transfer %f", account.Balance, tParams.Value))
		  }

		  account.Balance -= tParams.Value

      state, err = json.Marshal(account)
	    if err != nil {
		    return shim.Error(fmt.Sprintf("failed to serialize bank accountFrom %s information error %s", account.AccountNumber, err))
	    }

	    if err = stub.PutState(account.AccountNumber, []byte(state)); err != nil {
		    return shim.Error(fmt.Sprintf("failed to save bank accountFrom information with number %s, due to %s", account.AccountNumber, err))
	    }
    }

    b = true 

    if tParams.AccountTo == "0"{
      b = false
    }

    if b {
			state, err := stub.GetState(tParams.AccountTo)
      if err != nil {
        return shim.Error(fmt.Sprintf("failed to read accountTo %s information, due to %s", tParams.AccountTo, err))
      }
    
      if state == nil {
        return shim.Error(fmt.Sprintf("accountTo with account number %s doesn't exists", tParams.AccountTo))
      } 
      
      err = json.Unmarshal([]byte(state), &account)
	    if err != nil {
		    return shim.Error(fmt.Sprintf("failed to deserialize bank accountTo information error %s", err))
	    }

		  account.Balance += tParams.Value

      state, err = json.Marshal(account)
	    if err != nil {
		    return shim.Error(fmt.Sprintf("failed to serialize bank accountTo %s information error %s", account.AccountNumber, err))
	    }

	    if err = stub.PutState(account.AccountNumber, []byte(state)); err != nil {
		    return shim.Error(fmt.Sprintf("failed to save bank accountTo information with number %s, due to %s", account.AccountNumber, err))
	    }
		}

		return shim.Success(nil)
	},

	"accountHistory": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}
    
    unformattedHistory, err := stub.GetHistoryForKey(params [0])
	  if err != nil {
		  return shim.Error(fmt.Sprintf("failed to read history for account number %s, due to %s", params [0], err))
	  }
    
    var accountImage BankAccount
    buffer := bytes.Buffer{}
	  var prevPeriodBalance float64

	  prevPeriodBalance = 0.0

	  for unformattedHistory.HasNext() {
		  unformattedPeriod, err := unformattedHistory.Next()
		  if err != nil {
			  return shim.Error("failed Next() for HistoryQueryIteratorInterface")
		  }

		  err = json.Unmarshal(unformattedPeriod.Value, &accountImage)
		  if err != nil {
			  return shim.Error(fmt.Sprintf("failed to deserialize bank account information error %s", err))
		  }

		  delta := accountImage.Balance - prevPeriodBalance
		  if delta > 0.0 {
			  buffer.WriteString(fmt.Sprintf(" +%f", delta))
		  }

		  if delta < 0.0 {
			  buffer.WriteString(fmt.Sprintf(" %f", delta))
		  }

		prevPeriodBalance = accountImage.Balance
	}

	return shim.Success([]byte(buffer.String()))
	},
}

func (b *bankManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Bank Management chaincode is initialized")
	return shim.Success(nil)
}

func (b *bankManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
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
