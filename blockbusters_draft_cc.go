package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const MANUFACTURER = "MANUFACTURER"
const RETAILER = "RETAILER"
const CONSUMER = "CONSUMER"
const STATUS_VERIFIED = "VERIFIED"
const STATUS_IN_OWNERSHIP_TRANSIT = "IN OWNERSHIP TRANSIT"
const STATUS_POTENTIAL_COUNTERFEIT = "POTENTIAL COUNTERFEIT"
const STATUS_STOLEN = "STOLEN"
const STATUS_SOLD = "SOLD"
const STATUS_RETAILER = "RETAILER"
const TTYPE_CREATE = "CREATE"
const TTYPE_TRANSFER = "TRANSFER"
const TTYPE_CONFIRM = "CONFIRM"
const TTYPE_CHANGE_STATUS = "CHANGE STATUS"
const UNKNOWN_OWNER = "OWNER IS CURRENTLY UNKNOWN"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Item struct {
	Id           string        `json:"id"`
	CurrentOwner string        `json:"currentOwner"`
	Manufacturer string        `json:"manufacturer"`
	Barcode      string        `json:"barcode"`
	Status       string        `json:"status"`
	ScanCount    string        `json:"count"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	TType    string `json:"transactionType"`
	NewOwner string `json:"newOwner"`
	Location string `json:"location"`
	VDate    string `json:"vDate"`
}

type AllItems struct {
	Items []string `json:"items"`
}

type AllItemsDetails struct {
	Items []Item `json:"items"`
}

// ============================================================================================================================
// Init
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	var err error

	var items AllItems
	jsonAsBytes, _ := json.Marshal(items)
	err = stub.PutState("allItems", jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Run - Our entry point
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Run is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state
		return t.Init(stub, "init", args)
	} else if function == "transferOwnership" { //transfer ownership of item from current owner to new owner
		return t.transferOwnership(stub, args)
	} else if function == "changeStatus" { //transfer ownership of item from current owner to new owner
		return t.changeStatus(stub, args)
	} else if function == "createItem" { //create item with current user as the manufacturer
		return t.createItem(stub, args)
	} else if function == "confirmOwnership" { //confirm the transfer of ownership of an item to current user
		return t.confirmOwnership(stub, args)
	} else if function == "addScanCount" { //add scan count everytime an item is scanned after it has reached a retailer
		return t.addScanCount(stub, args)
	}

	fmt.Println("run did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - read a variable from chaincode state - (aka read)
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	if len(args) != 1 && len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments passed")
	}

	if function != "getItemDetailsWithID" && function != "getItemDetailsWithBarcode" && function != "getCurrentOwnerItems" && function != "getCurrentOwnerItemsByStatus" && function != "getCurrentOwnerItemsWithTxs" {
		return nil, errors.New("Invalid query function name.")
	}

	if function == "getItemDetailsWithID" {
		return t.getItemDetailsWithID(stub, args[0])
	}
	//	if function == "getItemDetailsWithBarcode" {
	//		return t.getItemDetailsWithBarcode(stub, args[0])
	//	}
	if function == "getCurrentOwnerItems" {
		return t.getCurrentOwnerItems(stub, args[0])
	}
	if function == "getCurrentOwnerItemsByStatus" {
		return t.getCurrentOwnerItemsByStatus(stub, args[0], args[1])
	}
	if function == "getCurrentOwnerItemsWithTxs" {
		return t.getCurrentOwnerItemsWithTxs(stub, args[0])
	}

	return nil, nil
}

// ============================================================================================================================
// Get Item Details using ID
// ============================================================================================================================
func (t *SimpleChaincode) getItemDetailsWithID(stub *shim.ChaincodeStub, itemId string) ([]byte, error) {

	fmt.Println("Start find Item using id")
	fmt.Println("Looking for Item #" + itemId)

	//get the item index
	iAsBytes, err := stub.GetState(itemId)
	if err != nil {
		return nil, errors.New("Failed to get Item #" + itemId)
	}

	return iAsBytes, nil

}

// ============================================================================================================================
// Get Item Details using Barcode
// ============================================================================================================================
//func (t *SimpleChaincode) getItemDetailsWithBarcode(stub *shim.ChaincodeStub, barcode string) ([]byte, error) {
//
//	fmt.Println("Start find Item using barcode")
//	fmt.Println("Looking for Item with barcode #" + barcode)
//
//	//get the AllItems index
//	allIAsBytes, err := stub.GetState("allItems")
//	if err != nil {
//		return nil, errors.New("Failed to get all Items")
//	}
//
//	var res AllItems
//	err = json.Unmarshal(allIAsBytes, &res)
//	if err != nil {
//		return nil, errors.New("Failed to Unmarshal all Items")
//	}
//
//	for i := range res.Items {
//
//		siAsBytes, err := stub.GetState(res.Items[i])
//		if err != nil {
//			return nil, errors.New("Failed to get Item")
//		}
//		var si Item
//		json.Unmarshal(siAsBytes, &si)
//
//		if si.Barcode == barcode {
//			return siAsBytes, nil
//		}
//
//	}
//
//	return nil, errors.New("Item with this specific barcode does not exist")
//
//}

// ============================================================================================================================
// Get All Items owned by a specific user (without transactions)
// ============================================================================================================================
func (t *SimpleChaincode) getCurrentOwnerItems(stub *shim.ChaincodeStub, user string) ([]byte, error) {

	fmt.Println("Start find getCurrentOwnerItems ")
	fmt.Println("Looking for All Items for " + user)

	//get the AllItems index
	allIAsBytes, err := stub.GetState("allItems")
	if err != nil {
		return nil, errors.New("Failed to get all Items")
	}

	var res AllItems
	err = json.Unmarshal(allIAsBytes, &res)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal all Items")
	}

	var rai AllItemsDetails

	for i := range res.Items {

		siAsBytes, err := stub.GetState(res.Items[i])
		if err != nil {
			return nil, errors.New("Failed to get Item")
		}
		var si Item
		json.Unmarshal(siAsBytes, &si)

		// get items without transactions
		if si.CurrentOwner == user || si.Manufacturer == user {
			si.Transactions = nil
			rai.Items = append(rai.Items, si)
		}

	}

	raiAsBytes, _ := json.Marshal(rai)

	return raiAsBytes, nil

}

// ============================================================================================================================
// Get All Items owned by a specific user and status (without transactions)
// ============================================================================================================================
func (t *SimpleChaincode) getCurrentOwnerItemsByStatus(stub *shim.ChaincodeStub, user string, status string) ([]byte, error) {

	fmt.Println("Start find getCurrentOwnerItemsByStatus ")
	fmt.Println("Looking for All Items for " + user + " with status: " + status)

	//get the AllItems index
	allIAsBytes, err := stub.GetState("allItems")
	if err != nil {
		return nil, errors.New("Failed to get all Items")
	}

	var res AllItems
	err = json.Unmarshal(allIAsBytes, &res)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal all Items")
	}

	var rai AllItemsDetails

	for i := range res.Items {

		siAsBytes, err := stub.GetState(res.Items[i])
		if err != nil {
			return nil, errors.New("Failed to get Item")
		}
		var si Item
		json.Unmarshal(siAsBytes, &si)

		// get items without transactions
		if (si.CurrentOwner == user) && (si.Status == status) {
			si.Transactions = nil
			rai.Items = append(rai.Items, si)
		}

	}

	raiAsBytes, _ := json.Marshal(rai)

	return raiAsBytes, nil

}

// ============================================================================================================================
// Get All Items Details for items owned by a specific user (including item's transactions)
// ============================================================================================================================
func (t *SimpleChaincode) getCurrentOwnerItemsWithTxs(stub *shim.ChaincodeStub, user string) ([]byte, error) {

	fmt.Println("Start find getCurrentOwnerItemsWithTxs ")
	fmt.Println("Looking for All Items Details for " + user)

	//get the AllItems index
	allIAsBytes, err := stub.GetState("allItems")
	if err != nil {
		return nil, errors.New("Failed to get all Items")
	}

	var res AllItems
	err = json.Unmarshal(allIAsBytes, &res)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal all Items")
	}

	var rai AllItemsDetails

	for i := range res.Items {

		siAsBytes, err := stub.GetState(res.Items[i])
		if err != nil {
			return nil, errors.New("Failed to get Item")
		}
		var si Item
		json.Unmarshal(siAsBytes, &si)

		if si.CurrentOwner == user || si.Manufacturer == user {
			rai.Items = append(rai.Items, si)
		}

	}

	raiAsBytes, _ := json.Marshal(rai)

	return raiAsBytes, nil

}

// ============================================================================================================================
// Create Items
// ============================================================================================================================
func (t *SimpleChaincode) createItem(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var err error
	fmt.Println("Running createItem")

	if len(args) != 5 {
		fmt.Println("Incorrect number of arguments. Expecting 6")
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	if args[1] != MANUFACTURER {
		fmt.Println("You are not allowed to create a new item")
		return nil, errors.New("You are not allowed to create a new item")
	}

	var it Item
	it.Id = args[0]
	it.Name = args[1]
	it.CurrentOwner = args[1]
	it.Manufacturer = args[1]
	it.Barcode = args[2]
	it.Status = STATUS_VERIFIED
	it.ScanCount = 0

	var tx Transaction
	tx.VDate = args[3]
	tx.Location = args[4]
	tx.TType = TTYPE_CREATE
	tx.NewOwner = it.CurrentOwner

	it.Transactions = append(it.Transactions, tx)

	//Commit item to ledger
	fmt.Println("createItem Commit Item To Ledger")
	itAsBytes, _ := json.Marshal(it)
	err = stub.PutState(it.Id, itAsBytes)
	if err != nil {
		return nil, err
	}

	//Update All Items Array
	allIAsBytes, err := stub.GetState("allItems")
	if err != nil {
		return nil, errors.New("Failed to get all Items")
	}
	var alli AllItems
	err = json.Unmarshal(allIAsBytes, &alli)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal all Items")
	}
	alli.Items = append(alli.Items, it.Id)

	allIuAsBytes, _ := json.Marshal(alli)
	err = stub.PutState("allItems", allIuAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Claim a item - can only be done by current owner
// ============================================================================================================================
func (t *SimpleChaincode) confirmOwnership(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var err error
	fmt.Println("Running confirmOwnership")

	if len(args) != 5 {
		fmt.Println("Incorrect number of arguments. Expecting 4 (ItemId, user, date, location, isRetailer)")
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	iAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get Item #" + args[0])
	}
	var itm Item
	err = json.Unmarshal(iAsBytes, &itm)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal Item #" + args[0])
	}

	if args[1] != itm.CurrentOwner {
		fmt.Println("You are not allowed to confirm ownership of this item")
		return nil, errors.New("You are not allowed to confirm ownership of this item")
	}

	if args[4] == "TRUE" {
		itm.Status = STATUS_VERIFIED
	} else {
		itm.Status = STATUS_RETAILER
	}

	var tx Transaction
	tx.VDate = args[2]
	tx.Location = args[3]
	tx.TType = TTYPE_CONFIRM
	tx.NewOwner = itm.CurrentOwner

	itm.Transactions = append(itm.Transactions, tx)

	//Commit updates item to ledger
	fmt.Println("confirmOwnership Commit Updates To Ledger")
	itAsBytes, _ := json.Marshal(itm)
	err = stub.PutState(itm.Id, itAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Increment scan counts - can only be done after the item has reached the retailers
// ============================================================================================================================
func (t *SimpleChaincode) addScanCount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	fmt.Println("Running addScanCount")

	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments. Expecting 1 (ItemId)")
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	//Update scan count
	iAsBytes, err := stub.getState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get Item#" + args[0])
	}
	var itm Item
	err = json.Unmarshal(iAsBytes, &itm)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal Item #" + args[0])
	}

	if itm.Status == STATUS_RETAILER {
		itm.ScanCount = itmScanCount + 1
	}

	return nil, nil
}

// ============================================================================================================================
// Transfer a item - can only be done by current owner
// ============================================================================================================================
func (t *SimpleChaincode) transferOwnership(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var err error
	fmt.Println("Running transferOwnership")

	if len(args) != 5 {
		fmt.Println("Incorrect number of arguments. Expecting 5 (ItemId, user, date, location, newOwner)")
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	//Update Item data
	iAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get Item #" + args[0])
	}
	var itm Item
	err = json.Unmarshal(iAsBytes, &itm)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal Item #" + args[0])
	}

	if args[1] != itm.CurrentOwner {
		return nil, errors.New("You are not allowed to transfer an item")
	}

	itm.CurrentOwner = args[4]
	itm.Status = STATUS_IN_OWNERSHIP_TRANSIT // There should be a big warning sign saying this item owner is unconfirmed

	var tx Transaction
	tx.VDate = args[2]
	tx.Location = args[3]
	tx.TType = TTYPE_TRANSFER
	tx.NewOwner = args[4]

	itm.Transactions = append(itm.Transactions, tx)

	//Commit updates item to ledger
	fmt.Println("transferOwnership Commit Updates To Ledger")
	itAsBytes, _ := json.Marshal(itm)
	err = stub.PutState(itm.Id, itAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Mark item with new status - can only be done by current owner and can only change to counterfeit, stolen or sold
// ============================================================================================================================
func (t *SimpleChaincode) changeStatus(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var err error
	fmt.Println("Running changeStatus")

	if len(args) != 5 {
		fmt.Println("Incorrect number of arguments. Expecting 5 (ItemId, user, date, location, newStatus)")
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	//Update Item data
	iAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get Item #" + args[0])
	}
	var itm Item
	err = json.Unmarshal(iAsBytes, &itm)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal Item #" + args[0])
	}

	if args[1] != itm.CurrentOwner {
		return nil, errors.New("You are not allowed to mark this item with new status")
	}

	if args[4] == STATUS_POTENTIAL_COUNTERFEIT || args[4] == STATUS_SOLD || args[4] == STATUS_STOLEN {
		itm.Status = args[4]
	} else {
		return nil, errors.New("Incorrect status change. New status must be counterfeit, stolen or sold.")
	}

	itm.CurrentOwner = UNKNOWN_OWNER

	var tx Transaction
	tx.VDate = args[2]
	tx.Location = args[3]
	tx.TType = TTYPE_CHANGE_STATUS
	tx.NewOwner = UNKNOWN_OWNER

	itm.Transactions = append(itm.Transactions, tx)

	//Commit updated item to ledger
	fmt.Println("changeStatus Commit Updates To Ledger")
	itAsBytes, _ := json.Marshal(itm)
	err = stub.PutState(itm.Id, itAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
