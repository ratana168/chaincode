package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type VoteStateValue struct {
	Status      string `json:"status"`
	Candidateid string `json:"candidateid"`
	Timestamp   string `json:"timestamp"`
	Ipaddr      string `json:"ipaddr"`
	Ua          string `json:"ua"`
	TxID        string `json:"txid"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "register" {
		return t.register(stub, args)
	} else if function == "stand" {
		return t.stand(stub, args)
	} else if function == "vote" {
		return t.vote(stub, args)
	} else if function == "unregister" {
		return t.unregister(stub, args)
	} else if function == "cancel" {
		return t.cancel(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {
		return t.read(stub, args)
	} else if function == "count" {
		return t.count(stub, args)
	} else if function == "failure" {
		return t.failure(stub, args)
	} else if function == "tokens" {
		return t.stateKeys(stub, []string{"vt_", args[0], args[1]})
	} else if function == "candidates" {
		return t.stateKeys(stub, []string{"cnt_", args[0], args[1]})
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// register (aVoteToken)  - register a vote token
func (t *SimpleChaincode) register(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running register()")
	var key, txid string
	var err error
	var jsonBytes []byte

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1. a vote token to be registered.")
	}

	key = "vt_" + args[0] // vt_<voteToken>
	txid = stub.GetTxID()
	val := VoteStateValue{Status: "NEW", Candidateid: "", Timestamp: "", Ipaddr: "", Ua: "", TxID: txid}
	jsonBytes, _ = json.Marshal(val)
	err = stub.PutState(key, jsonBytes) // add the vote token into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// stand (aCandidateId)  - stand as a candidate
func (t *SimpleChaincode) stand(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running stand()")
	var key, val string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1. a candidateid to be registered.")
	}

	key = "cnt_" + args[0] // cnt_<candidateId>
	val = "0"
	err = stub.PutState(key, []byte(val)) // add the vote token into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// vote (aVoteToken, aCandidateId, timestamp, ipaddr, ua)  - vote action
func (t *SimpleChaincode) vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running vote()")
	var key, keycnt, cntstr, txid string
	var jsonBytes, rb []byte
	val := new(VoteStateValue)

	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5 (votetoken, candidateid, timestamp, ipaddr, ua).")
	}

	keycnt = "cnt_" + args[1]
	rb, _ = t.count(stub, []string{args[1]})
	cntstr = string(rb)
	count, _ := strconv.Atoi(cntstr)
	key = "vt_" + args[0]
	jsonBytes, _ = t.read(stub, []string{args[0]})
	json.Unmarshal(jsonBytes, val)
	txid = stub.GetTxID()

	if val.Status == "NEW" {
		// the vote token exists and has not been voted.
		val.Status = "VOTED"
		val.Candidateid = args[1]
		val.Timestamp = args[2]
		val.Ipaddr = args[3]
		val.Ua = args[4]
		val.TxID = txid
		count = count + 1
		cntstr = strconv.Itoa(count)

		jsonBytes, _ = json.Marshal(val)
		stub.PutState(key, jsonBytes)
		stub.PutState(keycnt, []byte(cntstr))

	} else if val.Status == "VOTED" {
		stub.PutState("failure_"+args[0], []byte(txid))
		return nil, errors.New("DUPLICATED: the vote key has already been voted.")
	} else {
		stub.PutState("failure_"+args[0], []byte(txid))
		return nil, errors.New("ERROR")
	}
	return nil, nil
}

// read - query function to read a vote token status
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running read()")
	var key, ret string
	var err error
	var jsonBytes []byte
	val := new(VoteStateValue)

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting a vote token to query")
	}

	key = "vt_" + args[0]
	jsonBytes, err = stub.GetState(key)
	if err != nil {
		return nil, errors.New("Error occurred when getting state of " + key)
	}

	err = json.Unmarshal(jsonBytes, val)
	if err != nil {
		val.Status = "NA"
		val.Candidateid = ""
		val.Timestamp = ""
		val.Ipaddr = ""
		val.Ua = ""
		val.TxID = ""
	}
	jsonBytes, err = json.Marshal(val)
	ret = string(jsonBytes)
	if err != nil {
		ret = "{\"status\":\"ERROR\",\"candidateid\":\"\",\"timestamp\":\"\",\"ipaddr\":\"\",\"ua\":\"\",\"txid\":\"\"}"
	}
	return []byte(ret), nil
}

// count - query function to read a vote count
func (t *SimpleChaincode) count(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting a candidateid to query")
	}

	key = "cnt_" + args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		valAsbytes = []byte("-1")
	}

	return valAsbytes, nil
}

// failure - query function to return the transaction ID of the last failed vote
func (t *SimpleChaincode) failure(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting a vote token to be checked")
	}

	key = "failure_" + args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		valAsbytes = []byte("")
	}

	return valAsbytes, nil
}

// unregister (aVoteToken)  - unregister a vote token
func (t *SimpleChaincode) unregister(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running unregister()")
	var key string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1. a vote token to be unregistered.")
	}

	key = "vt_" + args[0] // vt_<voteToken>
	err = stub.DelState(key)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// cancel (aCandidateId)  - cancel a candidate
func (t *SimpleChaincode) cancel(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("running cancel()")
	var key string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1. a candidateid to be canceled.")
	}

	key = "cnt_" + args[0] // cnt_<candidateId>
	err = stub.DelState(key)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// stateKeys (prefix, startKey, endKey) - query function to list all stateKeys
func (t *SimpleChaincode) stateKeys(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var next, output string
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting prefix, start and end key.")
	}

	output = ""
	iter, err := stub.RangeQueryState(args[0]+args[1], args[0]+args[2])
	if err != nil {
		return nil, errors.New("Error occurred in RangeQueryState()")
	}
	for iter.HasNext() {
		next, _, _ = iter.Next()
		output = output + next + "\n"
	}

	return []byte(output), nil
}
