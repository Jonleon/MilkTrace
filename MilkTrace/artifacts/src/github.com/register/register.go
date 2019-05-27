package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("register")

//SimpleChaincode implementation
type SimpleChaincode struct {
}

//==================
//定义用户账户结构
//==================
type User struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Role     string `json:"role"`
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

	if function == "regist" {
		return t.regist(stub, args)
	}
	if function == "login" {
		return t.login(stub, args)
	}
	if function == "changePwd" {
		return t.changePwd(stub, args)
	}

	if function == "delete" {
		return t.delete(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}
	if function == "getHistoryForKey" {
		return t.getHistoryForKey(stub, args)
	}
	logger.Errorf("Unknown action, check the first argument, must be one of 'regist',delete', 'query', or 'transMoney'..."+
		" But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'regist',delete', 'query', or 'transMoney'..."+
		" But got: %v", args[0]))
}

//==================
//创建账号 args:ID | Password | Role
//==================
func (t *SimpleChaincode) regist(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Println("regist args:", args)
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 5,function followed by 1 accountID and 4 value")
	}

	var accountID string
	var Password, Role string //Password,Roel

	var err error
	accountID = args[0]
	Password = args[1]
	Role = args[2]

	var user User
	user.ID = accountID
	user.Password = Password
	user.Role = Role

	uBytes, _ := json.Marshal(user)
	//Get the state from the ledger
	Avalbytes, err := stub.GetState(accountID)
	fmt.Println(string(Avalbytes))
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes != nil {
		return shim.Error("this user already exist")
	}
	//Write the state back to the ledger
	err = stub.PutState(accountID, uBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(accountID + "账号创建成功！"))

}

//================
//验证账号密码是否匹配,登录 args:ID|Password
//================
func (t *SimpleChaincode) login(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	accountID := args[0]
	password := args[1]
	//query the ledger
	bytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if bytes == nil {
		return shim.Error("This account does not exists: " + accountID)
	}

	var user User
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}

	if user.Password == password {
		return shim.Success([]byte("correct password"))
	} else {
		return shim.Error("wrong password ")
	}
}

//================================
//查询账号 args:ID
//================================
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	var ID string // Entities
	var err error
	ID = args[0]
	// Get the state from the ledger
	Avalbytes, err := stub.GetState(ID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + ID + "\"}"
		return shim.Error(jsonResp)
	}
	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil count for " + ID + "\"}"
		return shim.Error(jsonResp)
	}
	jsonResp := string(Avalbytes)
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

//==============================
//更改用户密码 args:ID| OldPassword |newPassword
//==============================
func (t *SimpleChaincode) changePwd(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("ChangePassword")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//var account Account
	userID := args[0]
	userPassword := args[1]
	newPassword := args[2]
	var err error
	//query the ledger
	Bytes, _ := stub.GetState(userID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if Bytes == nil {
		return shim.Error("This accountt does not exists: " + userID)
	}
	var user User

	err = json.Unmarshal(Bytes, &user)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	//change password
	if user.Password == userPassword {
		user.Password = newPassword
	} else {
		return shim.Error("wrong password")
	}
	bytes, _ := json.Marshal(user)
	err = stub.PutState(userID, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("密码更改成功"))

}

//================
//删除账号 args：userID
//================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("Deleteusr")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountID := args[0]
	uBytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Fail to get user:" + err.Error())
	}
	if uBytes == nil {
		return shim.Error("this user is not found")
	}

	var user User
	err = json.Unmarshal(uBytes, &user)
	if err != nil {
		return shim.Error("Fail to get account:" + err.Error())
	}

	user = User{} //delete the user
	uBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(accountID, uBytes)

	err = stub.DelState(accountID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	err = stub.DelState(accountID)
	if err != nil {
		return shim.Success([]byte("delete sucessfully"))
	}

	return shim.Success([]byte("该账号已被删除"))
}

//==================================
//查看历史消息 args: userID
//==================================
func (t *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2,function followed by 1 accountID and 1 value")
	}

	var accountID string //Entities
	var err error
	accountID = args[0]
	//Get the state from the ledger
	//TODD:will be nice to have a GetAllState call to ledger
	HisInterface, err := stub.GetHistoryForKey(accountID)
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

//=================================
//main function
//=================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode:%s", err)
	}
}

