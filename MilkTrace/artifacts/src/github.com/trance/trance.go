package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("Service")

//SimpleChaincode implementation
type SimpleChaincode struct {
}

//==================
//牛奶的数据结构
//==================
type Milk struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"Price"`
	Produce_place string  `json:"produce_place"`
	Produce_time  string  `json:"produce_time"`
	Cow_info	  Cow	  `json:"cow_info"`
	Traces        []Trace `json:"traces"`
	Inner_traces  []Inner_trace `json:"inner_traces"`
}

//==========================
//追溯消息
//==========================
type Cow struct{
	Cow_type string `json:"cow_type"`
	Food	 string `json:"cow_food"`
	
}
type Inner_trace struct{
	Action		string `json:"action"`
	Time		string `json:"action_time"`
	Place		string `json:"place"`
	Worker		string `json:"worker"`
}
type Trace struct {
	TransID   string `json:"transID"`
	Place     string `json:"place"`
	TimeStamp string `json:"timeStamp"`
}

//安装Chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("######## Register Init ########")

	return shim.Success(nil)
}

//Invoke interface
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("######## Register Invoke ########")

	function, args := stub.GetFunctionAndParameters()

	if function == "milkInit" {
		return t.milkInit(stub, args)
	}
	if function == "trans" {
		return t.trans(stub, args)
	}

	if function == "query" {
		return t.query(stub, args)
	}
	if function == "getHistoryForKey" {
		return t.getHistoryForKey(stub, args)
	}
	if function == "testRangeQuery" {
		return t.testRangeQuery(stub, args)
	}
	logger.Errorf("Unknown action, check the first argument, must be one of 'milkInit', 'trans', "+
		"'query', 'testRangeQuery', But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'milkInit', 'trans', "+
		"'query', 'testRangeQuery', But got: %v", args[0]))
}

//==============================================================================================
//生产商权限下，发布一箱牛奶 milk args: ID |  Name | Price | produce_place | produce_time(不必传参)
//==============================================================================================
func (t *SimpleChaincode) milkInit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 4,function followed by 1 milkID and 3 value")
	}
	
	var milkID string        //Entities
	var Name string          //发布者编号、发布产品的名称
	var Price float64        //价格
	var produce_place string //产地
	var produce_time string  //生产时间
	var cow_type string
	var food string
	var err error
	
	milkID = args[0]
	Name = args[1]
	Price, _ = strconv.ParseFloat(args[2], 64)
	produce_place = args[3]
	cow_type = args[4]
	food = args[5]
	
	x := time.Now()
	produce_time = x.Format("2006-01-02 15:04:05")
	var milk Milk

	milk.ID = milkID
	milk.Name = Name
	milk.Price = Price
	milk.Produce_place = produce_place
	milk.Produce_time = produce_time
	milk.Cow_info.Cow_type = cow_type
	milk.Cow_info.Food = food
	uBytes, _ := json.Marshal(milk)
	//Get the state from the ledger
	//TODD:will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(milkID)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes != nil {
		return shim.Error("this id already exist")
	}

	//Write the state back to the ledger
	err = stub.PutState(milkID, uBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(milkID + "上链成功！"))
}

//============================================
//增加经销路径 args: milkID | agencyID | place
//============================================
func (t *SimpleChaincode) trans(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 3 {
		milkID := args[0]
		agencyID := args[1]
		place := args[2]
		mBytes, err := stub.GetState(milkID)
		if err != nil {
			shim.Error("Cann't get the milkID")
		}
		var milk Milk
		err = json.Unmarshal(mBytes, &milk)
		if err != nil {
			shim.Error("Cann't convert to MILK struct")
		}
		var trace Trace
		trace.TransID = agencyID
		c := time.Now()
		trace.TimeStamp = c.Format("2006-01-02 15:04:05")
		trace.Place = place
		milk.Traces = append(milk.Traces, trace)
		mBytes, _ = json.Marshal(milk)
		err = stub.PutState(milkID, mBytes)
		if err != nil {
			return shim.Error("fail to add trance:" + err.Error())
		}
	
		return shim.Success([]byte("添加运输溯源消息成功"))
	}
	if len(args) == 4{
		milkID := args[0]
		action := args[1]
		place := args[2]
		woker := args[3]
		mBytes, err := stub.GetState(milkID)
		if err != nil {
			shim.Error("Cann't get the milkID")
		}
		var milk Milk
		err = json.Unmarshal(mBytes, &milk)
		if err != nil {
			shim.Error("Cann't convert to MILK struct")
		}
		var inner_trace Inner_trace
		inner_trace.Action = action
		inner_trace.Place = place
		inner_trace.Worker = woker
		c := time.Now()
		inner_trace.Time = c.Format("2006-01-02 15:04:05")
		
		milk.Inner_traces = append(milk.Inner_traces, inner_trace)
		mBytes, _ = json.Marshal(milk)
		err = stub.PutState(milkID, mBytes)
		if err != nil {
			return shim.Error("fail to add trance:" + err.Error())
		}
	
		return shim.Success([]byte("添加内部溯源消息成功"))
	}
	return shim.Error("fail to add info")
}

//================================
//查询 args:milkID
//================================
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the Service to query")
	}
	var milkID string
	var err error
	milkID = args[0]
	// Get the state from the ledger
	Avalbytes, err := stub.GetState(milkID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + milkID + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil count for " + milkID + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := string(Avalbytes)
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

//===========================
//通过key查看历史记录 args:milkID
//===========================
func (t *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2,function followed by 1 accountID and 1 value")
	}

	var milkID string //Entities
	var err error
	milkID = args[0]
	//Get the state from the ledger
	//TODD:will be nice to have a GetAllState call to ledger
	HisInterface, err := stub.GetHistoryForKey(milkID)
	fmt.Println(HisInterface)
	Avalbytes, err := getHistoryListResult(HisInterface)
	if err != nil {
		return shim.Error("Failed to get history")
	}
	return shim.Success([]byte(Avalbytes))
}

func getHistoryListResult(resultsIterator shim.HistoryQueryIteratorInterface) ([]byte, error) {

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		item, _ := json.Marshal(queryResponse)
		buffer.Write(item)
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}

//==================================
//范围查询，args: 起始ID | 终止ID
//==================================
func (t *SimpleChaincode) testRangeQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//resultsIterator, err := stub.GetStateByRange("b1001", "b1010")
	startID := args[0]
	endID := args[1]
	resultsIterator, err := stub.GetStateByRange(startID, endID)
	if err != nil {
		return shim.Error("Query by Range failed")
	}
	services, err := getListResult(resultsIterator)
	if err != nil {
		return shim.Error("getListResult failed")
	}
	return shim.Success(services)
}

func getListResult(resultsIterator shim.StateQueryIteratorInterface) ([]byte, error) {

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}

//=================================
//main function
//=================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode:%s", err)
	}
}

