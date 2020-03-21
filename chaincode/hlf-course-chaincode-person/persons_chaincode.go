package hlf_course_chaincode_person

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Person structure to capture single person
type Person struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"second_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

var actions = map[string]func(stub shim.ChaincodeStubInterface, params []string) peer.Response{
	"addPerson": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of input parameters, expected 1, but got %d", len(params)))
		}

		// Define person variable
		var person Person
		if err := json.Unmarshal([]byte(params[0]), &person); err != nil {
			return shim.Error(fmt.Sprintf("failed to read input %s, error %s", params[0], err))
		}

		// check that newly added person is not exist
		personKey := fmt.Sprintf("%d", person.ID)
		state, err := stub.GetState(personKey)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read stat for key %s, error %s", person.ID, err))
		}

		if len(state) != 0 {
			return shim.Error(fmt.Sprintf("person with id = %s already exist", person.ID))
		}

		err = stub.PutState(personKey, []byte(params[0]))
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to store person with id = %s, due to %s", person.ID, err))
		}

		return shim.Success(nil)
	},

	"getPerson": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		state, err := stub.GetState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to read person information, id %s, due to %s", params[0], err))
		}
		
		if state == nil {
			return shim.Error(fmt.Sprintf("person with id %s doesn't exists", params[0]))
		}

		return shim.Success(state)
	},

	"delPerson": func(stub shim.ChaincodeStubInterface, params []string) peer.Response {
		if len(params) != 1 {
			return shim.Error(fmt.Sprintf("wrong number of parameters"))
		}

		err := stub.DelState(params[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to delete person information, id %s, due to %s", params[0], err))
		}

		return shim.Success(nil)
	},
}

// personManagement
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
