/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const (
	channelName      = "mychannel"
	chaincodeName    = "basic"
	mspOrg1          = "Org1MSP"
	org1UserID       = "appUser"
	fileSystemWallet = "wallet"
)

var basePath = filepath.Join(
	"..",
	"..",
	"test-network",
	"organizations",
	"peerOrganizations",
	"org1.example.com",
)

// build an in memory object with the network configuration (also known as a connection profile)
var ccpPath = filepath.Join(
	basePath,
	// "connection-org1.yaml", //Change this back to connection-org1.yaml once the Orderer Cert Info bug is fixed
	"connection-org1-withOrdererInfo.yaml",
)
var appUserCredPath = filepath.Join(
	basePath,
	"users",
	"User1@org1.example.com",
	"msp",
)

/**
 *  A test application to show basic queries operations with any of the asset-transfer-basic chaincodes
 *   -- How to submit a transaction
 *   -- How to query and check the results
 *
 * To see details of the GO SDK, Run the SDK in DEBUG mode.
 * Uncomment setLogLevel("DEBUG")
 */
func main() {

	log.Println("============ application-golang starts ============")

	// Set the Required Env for Debugging
	// setLogLevel("DEBUG")

	// Set Discovery as Local Host
	setDiscoveryAsLocalhost()

	// setup the wallet to hold the credentials of the application user
	wallet, err := gateway.NewFileSystemWallet(fileSystemWallet)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	fmt.Printf("AssetTransfer: Wallet Created in Filesystem\n")

	// Enroll Application User
	if !wallet.Exists(org1UserID) {
		err = populateWallet(wallet, appUserCredPath, org1UserID)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}
	something, _ := wallet.List()
	fmt.Printf("AssetTransfer: Wallet Updated for User: %v\n", something)

	// Create a new gateway instance to interact with the fabric network.
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, org1UserID),
		gateway.WithTimeout(100*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()
	fmt.Printf("AssetTransfer: Gateway connected with identity: %v\n", org1UserID)

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}
	fmt.Printf("AssetTransfer: Network Declared as: %v\n", network.Name())

	contract := network.GetContract(chaincodeName)
	fmt.Printf("AssetTransfer: Contract Declared as: %v\n", contract.Name())

	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	log.Println("--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")
	result, err = contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	printResultString(result)

	log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")
	log.Println("--> Submit Transaction: CreateAsset, ID: asset13, color: yellow, size: 5, owner: Tom, appraisalValue: 1300")
	result, err = contract.SubmitTransaction("CreateAsset", "asset13", "yellow", "5", "Tom", "1300")
	if err != nil {
		log.Printf("Failed to Submit transaction: %v", err)
	}

	log.Println("--> Evaluate Transaction: ReadAsset, function returns an asset with a given assetID")
	log.Println("--> Evaluate Transaction: ReadAsset, asset13")
	result, err = contract.EvaluateTransaction("ReadAsset", "asset13")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v\n", err)
	}
	printResultString(result)

	log.Println("--> Evaluate Transaction: AssetExists, function returns 'true' if an asset with given assetID exist")
	log.Println("--> Evaluate Transaction: AssetExists, asset1")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v\n", err)
	}
	printResultString(result)

	log.Println("--> Submit Transaction: TransferAsset Transfer asset to new owner")
	log.Println("--> Submit Transaction: TransferAsset Transfer asset: asset1 to new owner: Tom")
	_, err = contract.SubmitTransaction("TransferAsset", "asset1", "Tom")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	log.Println("--> Evaluate Transaction: ReadAsset, function returns asset attributes")
	log.Println("--> Evaluate Transaction: ReadAsset, 'asset1'")
	result, err = contract.EvaluateTransaction("ReadAsset", "asset1")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	printResultString(result)

	log.Println("============ application-golang ends ============")
}

func populateWallet(wallet *gateway.Wallet, credPath string, user string) error {

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")

	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put(user, identity)
}

func setLogLevel(sdkLogLevel string) {

	err := os.Setenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL", sdkLogLevel)
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}
	fmt.Printf("AssetTransfer: Setting FABRIC_SDK_CLIENT_LOGGING_LEVEL to: %v\n", os.Getenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL"))
}

func setDiscoveryAsLocalhost() {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}
	fmt.Printf("AssetTransfer: Setting DISCOVERY_AS_LOCALHOST to: %v\n", os.Getenv("DISCOVERY_AS_LOCALHOST"))
}

func printResultString(result []byte) {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, result, "", "  ")
	if error != nil {
		log.Println("AssetTransfer: Unable to Pretty Print JSON - parse error: ", error)
		log.Println(string(result))
	} else {
		log.Println(string(prettyJSON.Bytes()))
	}
}
