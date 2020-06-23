package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type CreditCard struct {
	CardNumber     string `json:"card_number"`
	ExpirationDate string `json:"expiration_date"`
	CVV            string `json:"CVV_code"`
	PersonID       string `json:"person_id"`
	AccountNumber  string `json:"account_number"`
}

type BankAccount struct {
	PersonID      string  `json:"person_id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
}

type Person struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"second_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type cardManagement struct {
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addCreditCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of arguments"))
		}

		var card CreditCard
		err := json.Unmarshal([]byte(params[0]), &card)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to desirialize bank account information error %s", err))
		}

		// Need to check whenever card.PersonID is exists
		response := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(card.PersonID)}, "mychannel")
		if response.Status == shim.ERROR {
			_, err := json.Marshal(response)
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
			}
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s, due to %s", card.PersonID, response.Message))
		}

		// Need to check whenever card.AccountNumber is exists
		response = stub.InvokeChaincode("bank_chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			_, err = json.Marshal(response)
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
			}
			return shim.Error(fmt.Sprintf("failed to create credit card for bank account with number %s, due to %s", card.AccountNumber, response.Message))
		}

		cardState, err := stub.GetState(card.CardNumber)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to create credit card due to %s", err))
		}

		if cardState != nil {
			return shim.Error(fmt.Sprintf("credit card with number %s already exists", card.CardNumber))
		}

		if err := stub.PutState(card.CardNumber, []byte(params[0])); err != nil {
			shim.Error(fmt.Sprintf("failed to save credit card with number %s, due to %s", card.CardNumber, err))
		}

		return shim.Success(nil)
	},

	"getCreditCardInfo": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read information of credit card with number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("credit card with number %s doesn't exists", params[0]))
		}
		var card CreditCard
		if err := json.Unmarshal([]byte(state), &card); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", state, err))
		}
		cardInfo := fmt.Sprintf("Expiration date: %s, CVV-code: %s", card.ExpirationDate, card.CVV)

		response := stub.InvokeChaincode("persons_chaincode", [][]byte{[]byte("getPerson"), []byte(card.PersonID)}, "mychannel")
		if response.Status == shim.ERROR {
			_, err := json.Marshal(response)
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
			}
			return shim.Error(fmt.Sprintf("failed to create credit card for person with id %s, due to %s", card.PersonID, response.Message))
		}
		var person Person
		if err := json.Unmarshal([]byte(response.GetPayload()), &person); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", response.GetPayload(), err))
		}
		personInfo := fmt.Sprintf("ID: %s, FirstName: %s, LastName: %s, Address: %s, Phone: %s", person.ID, person.FirstName, person.LastName, person.Address, person.Phone)

		response = stub.InvokeChaincode("bank_chaincode", [][]byte{[]byte("getBalance"), []byte(card.AccountNumber)}, "mychannel")
		if response.Status == shim.ERROR {
			_, err = json.Marshal(response)
			if err != nil {
				return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
			}
			return shim.Error(fmt.Sprintf("failed to create credit card for bank account with number %s, due to %s", card.AccountNumber, response.Message))
		}
		var account BankAccount
		if err := json.Unmarshal([]byte(response.GetPayload()), &account.Balance); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", response.GetPayload(), err))
		}
		accountInfo := fmt.Sprintf("Balance: %f, ", account.Balance) + fmt.Sprintf("Bank account number: %s", card.AccountNumber)

		return shim.Success([]byte(fmt.Sprintf("credit card information with number %s: Person Info - %s , Bank account Info - %s, Credit card Info: %s", params[0], personInfo, accountInfo, cardInfo)))
	},

	"delCreditCard": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read credit card information with number %s, due to %s", params[0], err))
		}

		if state == nil {
			return shim.Error(fmt.Sprintf("credit card with number %s doesn't exists", params[0]))
		}

		err = stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete credit card information with number %s, due to %s", params[0], err))
		}
		return shim.Success(nil)
	},
}

func (b cardManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Credit card Management chaincode is initialized")
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
