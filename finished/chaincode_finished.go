/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"
	"log"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var openTradesStr = "_opentrades"				//name for the key/value that will store all open trades


var invoiceIndexStr = "_invoiceindex" 
var accountIndexStr = "_accountindex"
var paymentIndexStr = "_paymentindex"

//for invoice
type Invoice struct{
	VendorID string `json:"vendorid"`
	CustomerID string `json:"customerid"`
	InvoiceNumber string `json:"invoicenumber"`	
	InvoiceAmount float64 `json:"invoiceamount"`
	Currency string `json:"currency"`	
	Material string `json:"material"`
	Quantity int `json:"quantity"`
	TradeID string `json:"tradeid"`
	PaymentDate string `json:"paymentdate"`
	Status string `json:"status"`
	NewPaymentDate string `json:"newpaymentdate"`
} 

//for account
type Invoice struct{
	ID string `json:"vendorid"`
	AccountName string `json:"accountname"`
	AccountType string `json:"accounttype"`	
	Address string `json:"address"`
	BankAccountNumber int `json:"bankaccountnumber"`	
	Phone string `json:"phone"`
	BankerID string `json:"bankerid"`

} 

//for payment
type Invoice struct{
	PaymentID string `json:"paymentId"`
	VendorID string `json:"vendorid"`
	CustomerID string `json:"customerid"`
	InvoiceID string `json:"invoiceid"`	
	Amount float64 `json:"amount"`
	Currency string `json:"currency"`
	BankerID string `json:"bankerid"	
	PaymentDate string `json:"paymentdate"`
	TradeID string `json:"tradeid"`
	NewPaymentDate string `json:"newpaymentdate"`
} 

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error
	fmt.Printf("intot init")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(invoiceIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	var emptyInv []string
	jsonAsBytesInv, _ := json.Marshal(emptyInv)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(invoiceIndexStr, jsonAsBytesInv)
	if err != nil {
		return nil, err
	}
	
	var trades AllTrades
	jsonAsBytes, _ = json.Marshal(trades)								//clear the open trade struct
	err = stub.PutState(openTradesStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "create_invoice" {									//create a new invoice
		return t.create_invoice(stub, args)
	else if function == "create_account" {									//create a new account
		return t.create_invoice(stub, args)
	else if function == "create_payment" {									//create a new payment
		return t.create_invoice(stub, args)
	} else if function == "set_user" {										//change owner of a invoice
		res, err := t.set_user(stub, args)
		cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "open_trade" {									//create a new trade order
		return t.open_trade(stub, args)
	} else if function == "perform_trade" {									//forfill an open trade order
		res, err := t.perform_trade(stub, args)
		cleanTrades(stub)													//lets clean just in case
		return res, err
	} else if function == "remove_trade" {									//cancel an open trade order
		return t.remove_trade(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}



// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}


//this is for invoice
func (t *SimpleChaincode) create_invoice(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	if len(args) != 11 {
		return nil, errors.New("Incorrect number of arguments. Expecting 11")
	}
	
	//input sanitation
	fmt.Println("- start init invoice")
	if len(args[0]) <= 0 {
		// Vendor ID //
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		// Customer ID //
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		//Invoice Number //
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		// Invoice Amount //
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		// Currency //
		return nil, errors.New("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		// Material //
		return nil, errors.New("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		// Quantity //
		return nil, errors.New("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		// Trader ID //
		return nil, errors.New("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		// Payment Date //
		return nil, errors.New("9th argument must be a non-empty string")
	}
	if len(args[9]) <= 0 {
		// Status //
		return nil, errors.New("10th argument must be a non-empty string")
	}
	if len(args[10]) <= 0 {
		// New Payment Date //
		return nil, errors.New("11th argument must be a non-empty string")
	}
	VendorID := args[0]
	CustomerID := args[1]
	InvoiceNumber := args[2]
	InvoiceAmount := args[3]
	Currency := args[4]
	Material := args[5]
	Quantity := args[6]
	TraderID := args[7]
	PaymentDate := args[8]
	Status := args[9]
	NewPaymentDate := args[10]

	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	//check if invoice already exists
	invoiceAsBytes, err := stub.GetState(InvoiceNumber)
	if err != nil {
		return nil, errors.New("Failed to get Invoice number")
	}
	res := Invoice{}
	json.Unmarshal(invoiceAsBytes, &res)
	if res.InvoiceNumber == InvoiceNumber{
		fmt.Println("This invoice arleady exists: " + InvoiceNumber)
		fmt.Println(res);
		return nil, errors.New("This invoice arleady exists")				//all stop a invoice by this name exists
	}
	
	

	str := `{"vendorid": "` + VendorID + `", "customerid": "` + CustomerID + `", "invoicenumber": "` + InvoiceNumber + `", "invoiceamount": ` + InvoiceAmount + `, "currency": "` + Currency + `", "material": "` + Material + `","quantity":` + Quantity + `,"traderid":"` + TraderID + `","paymentdate":"` + PaymentDate + `","Status":"` + Status + `","newpaymentdate":"` + NewPaymentDate + `"}`
	err = stub.PutState(VendorID, []byte(str))									//store invoice with id as key
	if err != nil {
		return nil, err
	}
		
	//get the invoice index
	invoicesAsBytes, err := stub.GetState(invoiceIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get invoice index")
	}
	var invoiceIndex []string
	json.Unmarshal(invoicesAsBytes, &invoiceIndex)							//un stringify it aka JSON.parse()
	
	//append
	invoiceIndex = append(invoiceIndex, VendorID)									//add invoice name to index list
	fmt.Println("! invoice index: ", invoiceIndex)
	jsonAsBytes, _ := json.Marshal(invoiceIndex)
	err = stub.PutState(invoiceIndexStr, jsonAsBytes)						//store name of invoice

	fmt.Println("- end init invoice")
	return nil, nil
} 

//this is for account

	 string `json:"vendorid"`
	 string `json:"accountname"`
	 string `json:"accounttype"`	
	 string `json:"address"`
	  int `json:"bankaccount"`	
	 string `json:"phone"`
	 string `json:"bankerid"`
	
func (t *SimpleChaincode) create_account(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	
	ID := args[0]
	AccountName := args[1]
	AccountType := args[2]
	InvoiceAmount := args[3]
	Address := args[4]
	BankAccountNumber := args[5]
	Phone := args[6]
	BankerID := args[7]

	

	//check if account already exists
	accountAsBytes, err := stub.GetState(ID)
	if err != nil {
		return nil, errors.New("Failed to get account")
	}
	res := Account{}
	json.Unmarshal(accountAsBytes, &res)
	if res.ID == ID{
		fmt.Println("This invoice arleady exists: " + ID)
		fmt.Println(res);
		return nil, errors.New("This account arleady exists")				
	}
	
	
	str := `{"ID": "` + ID + `", "AccountName": "` + AccountName + `", "AccountType": "` + AccountType + `", "InvoiceAmount": ` + InvoiceAmount + `, "Address": "` + Address + `", "BankAccountNumber": "` + BankAccountNumber + `","Phone":` + Phone + `,"traderid":"` + TraderID + `","BankerID":"` + BankerID + `"}`
	
	err = stub.PutState(ID, []byte(str))									
	if err != nil {
		return nil, err
	}
		
	//get the account index
	accountAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get account index")
	}
	var invoiceIndex []string
	json.Unmarshal(invoicesAsBytes, &invoiceIndex)							//un stringify it aka JSON.parse()
	
	//append
	accountIndex = append(accountIndex, ID)									//add account id to index list
	fmt.Println("! invoice index: ", accountAsBytes)
	jsonAsBytes, _ := json.Marshal(accountAsBytes)
	err = stub.PutState(accountIndexStr, jsonAsBytes)						//store name of account

	fmt.Println("- end init account")
	return nil, nil
} 
//create payment
func (t *SimpleChaincode) create_payment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	PaymentID := args[0]
	VendorID := args[1]
	CustomerID := args[2]
	InvoiceID := args[3]
	Amount := args[4]
	Currency := args[5]
	BankerID := args[6]
	PaymentDate := args[7]
	TraderID := args[8]
	NewPaymentDate := args[9]

	

	//check if payment already exists
	paymentAsBytes, err := stub.GetState(ID)
	if err != nil {
		return nil, errors.New("Failed to get payment")
	}
	res := Payment{}
	json.Unmarshal(paymentAsBytes, &res)
	if res.ID == ID{
		fmt.Println("This invoice arleady exists: " + ID)
		fmt.Println(res);
		return nil, errors.New("This account arleady exists")				
	}
	
	
	str := `{"vendorid": "` + VendorID + `", "customerid": "` + CustomerID + `", "InvoiceID": "` + InvoiceID + `", "Amount": ` + Amount + `, "currency": "` + Currency + `", "BankerID": "` + BankerID + `","paymentdate":"` + PaymentDate + `","traderid":"` + TraderID + `","newpaymentdate":"` + NewPaymentDate + `"}`
	
	err = stub.PutState(ID, []byte(str))									
	if err != nil {
		return nil, err
	}
		
	//get the account index
	paymentAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get account index")
	}
	var paymentIndex []string
	json.Unmarshal(invoicesAsBytes, &paymentIndex)							//un stringify it aka JSON.parse()
	
	//append
	paymentIndex = append(paymentIndex, ID)									//add account id to index list
	fmt.Println("! invoice index: ", paymentAsBytes)
	jsonAsBytes, _ := json.Marshal(paymentAsBytes)
	err = stub.PutState(accountIndexStr, jsonAsBytes)						//store name of account

	fmt.Println("- end init account")
	return nil, nil
} 
// ============================================================================================================================
// Set User Permission on invoice
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	invoiceAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := invoice{}
	json.Unmarshal(invoiceAsBytes, &res)										//un stringify it aka JSON.parse()
	res.User = args[1]														//change the user
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the invoice with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set user")
	return nil, nil
}

// ============================================================================================================================
// Open Trade - create an open trade for a invoice you want with invoices you have 
// ============================================================================================================================
func (t *SimpleChaincode) open_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var will_size int
	var trade_away Description
	
	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}
	if len(args)%2 == 0{
		return nil, errors.New("Incorrect number of arguments. Expecting an odd number")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	open := AnOpenTrade{}
	open.User = args[0]
	open.Timestamp = makeTimestamp()											//use timestamp as an ID
	open.Want.Color = args[1]
	open.Want.Size =  size1
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	for i:=3; i < len(args); i++ {												//create and append each willing trade
		will_size, err = strconv.Atoi(args[i + 1])
		if err != nil {
			msg := "is not a numeric string " + args[i + 1]
			fmt.Println(msg)
			return nil, errors.New(msg)
		}
		
		trade_away = Description{}
		trade_away.Color = args[i]
		trade_away.Size =  will_size
		fmt.Println("! created trade_away: " + args[i])
		jsonAsBytes, _ = json.Marshal(trade_away)
		err = stub.PutState("_debug2", jsonAsBytes)
		
		open.Willing = append(open.Willing, trade_away)
		fmt.Println("! appended willing to open")
		i++;
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()
	
	trades.OpenTrades = append(trades.OpenTrades, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(openTradesStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}

// ============================================================================================================================
// Perform Trade - close an open trade and move ownership
// ============================================================================================================================
func (t *SimpleChaincode) perform_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0		1					2					3				4					5
	//[data.id, data.closer.user, data.closer.name, data.opener.user, data.opener.color, data.opener.size]
	if len(args) < 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}
	
	fmt.Println("- start close trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	size, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, errors.New("6th argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)															//un stringify it aka JSON.parse()
	
	for i := range trades.OpenTrades{																//look for the trade
		fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			
			
			invoiceAsBytes, err := stub.GetState(args[2])
			if err != nil {
				return nil, errors.New("Failed to get thing")
			}
			closersinvoice := invoice{}
			json.Unmarshal(invoiceAsBytes, &closersinvoice)											//un stringify it aka JSON.parse()
			
			//verify if invoice meets trade requirements
			if closersinvoice.Color != trades.OpenTrades[i].Want.Color || closersinvoice.Size != trades.OpenTrades[i].Want.Size {
				msg := "invoice in input does not meet trade requriements"
				fmt.Println(msg)
				return nil, errors.New(msg)
			}
			
			invoice, e := findinvoice4Trade(stub, trades.OpenTrades[i].User, args[4], size)			//find a invoice that is suitable from opener
			if(e == nil){
				fmt.Println("! no errors, proceeding")

				t.set_user(stub, []string{args[2], trades.OpenTrades[i].User})						//change owner of selected invoice, closer -> opener
				t.set_user(stub, []string{invoice.Name, args[1]})									//change owner of selected invoice, opener -> closer
			
				trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)		//remove trade
				jsonAsBytes, _ := json.Marshal(trades)
				err = stub.PutState(openTradesStr, jsonAsBytes)										//rewrite open orders
				if err != nil {
					return nil, err
				}
			}
		}
	}
	fmt.Println("- end close trade")
	return nil, nil
}

// ============================================================================================================================
// findinvoice4Trade - look for a matching invoice that this user owns and return it
// ============================================================================================================================
func findinvoice4Trade(stub shim.ChaincodeStubInterface, user string, color string, size int )(m invoice, err error){
	var fail invoice;
	fmt.Println("- start find invoice 4 trade")
	fmt.Println("looking for " + user + ", " + color + ", " + strconv.Itoa(size));

	//get the invoice index
	invoicesAsBytes, err := stub.GetState(invoiceIndexStr)
	if err != nil {
		return fail, errors.New("Failed to get invoice index")
	}
	var invoiceIndex []string
	json.Unmarshal(invoicesAsBytes, &invoiceIndex)								//un stringify it aka JSON.parse()
	
	for i:= range invoiceIndex{													//iter through all the invoices
		//fmt.Println("looking @ invoice name: " + invoiceIndex[i]);

		invoiceAsBytes, err := stub.GetState(invoiceIndex[i])						//grab this invoice
		if err != nil {
			return fail, errors.New("Failed to get invoice")
		}
		res := invoice{}
		json.Unmarshal(invoiceAsBytes, &res)										//un stringify it aka JSON.parse()
		//fmt.Println("looking @ " + res.User + ", " + res.Color + ", " + strconv.Itoa(res.Size));
		
		//check for user && color && size
		if strings.ToLower(res.User) == strings.ToLower(user) && strings.ToLower(res.Color) == strings.ToLower(color) && res.Size == size{
			fmt.Println("found a invoice: " + res.Name)
			fmt.Println("! end find invoice 4 trade")
			return res, nil
		}
	}
	
	fmt.Println("- end find invoice 4 trade - error")
	return fail, errors.New("Did not find invoice to use in this trade")
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// ============================================================================================================================
// Remove Open Trade - close an open trade
// ============================================================================================================================
func (t *SimpleChaincode) remove_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0
	//[data.id]
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																//un stringify it aka JSON.parse()
	
	for i := range trades.OpenTrades{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)				//remove this trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(openTradesStr, jsonAsBytes)												//rewrite open orders
			if err != nil {
				return nil, err
			}
			break
		}
	}
	
	fmt.Println("- end remove trade")
	return nil, nil
}

// ============================================================================================================================
// Clean Up Open Trades - make sure open trades are still possible, remove choices that are no longer possible, remove trades that have no valid choices
// ============================================================================================================================
func cleanTrades(stub shim.ChaincodeStubInterface)(err error){
	var didWork = false
	fmt.Println("- start clean trades")
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																		//un stringify it aka JSON.parse()
	
	fmt.Println("# trades " + strconv.Itoa(len(trades.OpenTrades)))
	for i:=0; i<len(trades.OpenTrades); {																		//iter over all the known open trades
		fmt.Println(strconv.Itoa(i) + ": looking at trade " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10))
		
		fmt.Println("# options " + strconv.Itoa(len(trades.OpenTrades[i].Willing)))
		for x:=0; x<len(trades.OpenTrades[i].Willing); {														//find a invoice that is suitable
			fmt.Println("! on next option " + strconv.Itoa(i) + ":" + strconv.Itoa(x))
			_, e := findinvoice4Trade(stub, trades.OpenTrades[i].User, trades.OpenTrades[i].Willing[x].Color, trades.OpenTrades[i].Willing[x].Size)
			if(e != nil){
				fmt.Println("! errors with this option, removing option")
				didWork = true
				trades.OpenTrades[i].Willing = append(trades.OpenTrades[i].Willing[:x], trades.OpenTrades[i].Willing[x+1:]...)	//remove this option
				x--;
			}else{
				fmt.Println("! this option is fine")
			}
			
			x++
			fmt.Println("! x:" + strconv.Itoa(x))
			if x >= len(trades.OpenTrades[i].Willing) {														//things might have shifted, recalcuate
				break
			}
		}
		
		if len(trades.OpenTrades[i].Willing) == 0 {
			fmt.Println("! no more options for this trade, removing trade")
			didWork = true
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)					//remove this trade
			i--;
			
		}
		
		i++
		fmt.Println("! i:" + strconv.Itoa(i))
		if i >= len(trades.OpenTrades) {																	//things might have shifted, recalcuate
			break
		}
	}

	if(didWork){
		fmt.Println("! saving open trade changes")
		jsonAsBytes, _ := json.Marshal(trades)
		err = stub.PutState(openTradesStr, jsonAsBytes)														//rewrite open orders
		if err != nil {
			return err
		}
	}else{
		fmt.Println("! all open trades are fine")
	}

	fmt.Println("- end clean trades")
	return nil
}