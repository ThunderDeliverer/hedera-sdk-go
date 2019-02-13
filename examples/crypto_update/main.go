package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hashgraph/hedera-sdk-go"
)

func main() {
	//
	// Generate keys
	//

	// Read and decode the operator secret key
	operatorSecret, err := hedera.SecretKeyFromString(os.Getenv("OPERATOR_SECRET"))
	if err != nil {
		panic(err)
	}

	// Generate a new keypair for the new account
	secret, _ := hedera.GenerateSecretKey()
	public := secret.Public()

	fmt.Printf("secret = %v\n", secret)
	fmt.Printf("public = %v\n", public)

	//
	// Connect to Hedera
	//

	client, err := hedera.Dial("testnet.hedera.com:50003")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	//
	// Send transaction to create account
	//

	nodeAccountID := hedera.AccountID{Account: 3}
	operatorAccountID := hedera.AccountID{Account: 2}
	response, err := client.CreateAccount().
		Key(public).
		InitialBalance(0).
		Operator(operatorAccountID).
		Node(nodeAccountID).
		Memo("[test] hedera-sdk-go v2").
		Sign(operatorSecret).
		Execute()

	if err != nil {
		panic(err)
	}

	transactionID := response.ID
	fmt.Printf("created account; transaction = %v\n", transactionID)

	//
	// Get receipt to prove we created it ok
	//

	fmt.Printf("wait for 2s...\n")
	time.Sleep(2 * time.Second)

	receipt, err := client.Transaction(*transactionID).Receipt().Get()
	if err != nil {
		panic(err)
	}

	if receipt.Status != hedera.StatusSuccess {
		panic(fmt.Errorf("transaction has a non-successful status: %v", receipt.Status.String()))
	}

	fmt.Printf("account = %v\n", *receipt.AccountID)

	// now update the account
	operatorAccountID = *receipt.AccountID

	expiryTime := time.Now().AddDate(0, 0, 365) // add one day

	fmt.Printf("updating account = %v\n", *receipt.AccountID)
	response, err = client.CryptoUpdate(*receipt.AccountID).
		ExpiresAt(expiryTime).
		AutoRenewPeriod(36000, 0).
		Operator(operatorAccountID).
		SendRecordThreshold(100000000000000).
		ReceiveRecordThreshold(100000000000000).
		Sign(secret).
		Execute()
}
