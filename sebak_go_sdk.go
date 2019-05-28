package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	"boscoin.io/sebak/lib/block"
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/common/keypair"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/transaction"
	"boscoin.io/sebak/lib/transaction/operation"
)

func main() {

	var targetAccount keypair.KP
	var tx transaction.Transaction
	var connection *common.HTTP2Client
	var sourceAccount block.BlockAccount
	var err error
	var endpoint *common.Endpoint
	var amount common.Amount
	var sender keypair.KP
	var seed string

	// TestNet network ID
	var networkID = "sebak-test-network"
	// TestNet endpoint
	var strEndpoint = "https://testnet-sebak.blockchainos.org:443"

	// Source account secrete seed
	seed = os.Args[1]

	// Receiver account's public address
	if targetAccount, err = keypair.Parse(os.Args[2]); err != nil {
		log.Fatal("Target account's public address : ", err)
		os.Exit(1)
	}

	// In this example, amount is 0.1 BOS.
	// To create account, amount should be bigger than 0.1 BOS.
	if amount, err = cmdcommon.ParseAmountFromString("1000000"); err != nil {
		log.Fatal("amount : ", err)
		os.Exit(1)
	}

	if connection, err = common.NewHTTP2Client(0, 0, true); err != nil {
		log.Fatal("Error while creating network client : ", err)
		os.Exit(1)
	}
	if endpoint, err = common.ParseEndpoint(strEndpoint); err != nil {
		log.Fatal("endpoint : ", err)
	}
	client := network.NewHTTP2NetworkClient(endpoint, connection)

	if sender, err = keypair.Parse(seed); err != nil {
		log.Fatal("Source account secret seed : ", err)
	}

	// Get the BlockAccount of the sender
	if sourceAccount, err = getAccountDetails(client, sender); err != nil {
		log.Fatal("Could not fetch source account : ", err)
		os.Exit(1)
	}

	// Check that source account's balance is enough before sending the transaction
	{
		fee := common.BaseFee

		_, err = sourceAccount.GetBalance().Sub(amount + fee)
		if err != nil {
			fmt.Printf("Attempting to draft %v GON (+ %v fees), but source account only have %v GON\n",
				amount, fee, sourceAccount.GetBalance())
			os.Exit(1)
		}
	}

	if os.Args[3] == "create" {
		tx = MakeTransactionCreateAccount(sender, targetAccount, amount, sourceAccount.SequenceID)
	}
	if os.Args[3] == "payment" {
		tx = MakeTransactionPayment(sender, targetAccount, amount, sourceAccount.SequenceID)
	}

	tx.Sign(sender, []byte(networkID))

	fmt.Println(tx)

	// Send request
	var responseBody []byte
	if responseBody, err = client.SendTransaction(tx); err != nil {
		log.Fatal("Network error: ", err, " body : ", string(responseBody))
		os.Exit(1)
	}

	// Check target account balance
	time.Sleep(5 * time.Second)
	if recv, err := getAccountDetails(client, targetAccount); err != nil {
		fmt.Println("Account ", targetAccount.Address(), " did not appear after 5 seconds")
	} else {
		fmt.Println("Target account after 5 seconds : ", recv)
	}

}

func MakeTransactionCreateAccount(kpSource keypair.KP, kpDest keypair.KP, amount common.Amount, seqid uint64) transaction.Transaction {
	var opb operation.CreateAccount
	var fee common.Amount
	opb = operation.NewCreateAccount(kpDest.Address(), amount, "")
	fee = common.BaseFee

	op := operation.Operation{
		H: operation.Header{
			Type: operation.TypeCreateAccount,
		},
		B: opb,
	}

	txBody := transaction.Body{
		Source:     kpSource.Address(),
		Fee:        fee,
		SequenceID: seqid,
		Operations: []operation.Operation{op},
	}

	tx := transaction.Transaction{
		H: transaction.Header{
			Version: common.TransactionVersionV1,
			Created: common.NowISO8601(),
			Hash:    txBody.MakeHashString(),
		},
		B: txBody,
	}

	return tx
}

func MakeTransactionPayment(kpSource keypair.KP, kpDest keypair.KP, amount common.Amount, seqid uint64) transaction.Transaction {
	opb := operation.NewPayment(kpDest.Address(), amount)

	op := operation.Operation{
		H: operation.Header{
			Type: operation.TypePayment,
		},
		B: opb,
	}

	txBody := transaction.Body{
		Source:     kpSource.Address(),
		Fee:        common.BaseFee,
		SequenceID: seqid,
		Operations: []operation.Operation{op},
	}

	tx := transaction.Transaction{
		H: transaction.Header{
			Version: common.TransactionVersionV1,
			Created: common.NowISO8601(),
			Hash:    txBody.MakeHashString(),
		},
		B: txBody,
	}

	return tx
}

func getAccountDetails(conn *network.HTTP2NetworkClient, sender keypair.KP) (block.BlockAccount, error) {
	var ba block.BlockAccount
	var err error
	var retBody []byte

	if retBody, err = conn.Get("/api/v1/accounts/" + sender.Address()); err != nil {
		return ba, err
	}

	err = json.Unmarshal(retBody, &ba)
	return ba, err
}
