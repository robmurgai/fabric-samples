/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {

	log.Println("============ application-golang starts ============")

	// log.Printf("\nRob-Debug: Environemnt Vars total: %v\n", len(os.Environ()))
	// for _, envs := range os.Environ() {
	// 	log.Printf("\nRob-Debug: Environemnt Var is: %v\n", envs)
	// }

	hfcset := "{\"debug\":\"console\"}"
	err := os.Setenv("HFC_LOGGING", hfcset)
	if err != nil {
		log.Fatalf("Error setting HFC_LOGGING environemnt variable: %v", err)
	}

	err = os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	// key := "localhostEnvVarName"
	// val, ok := os.LookupEnv(key)
	// if !ok {
	// 	fmt.Printf("Rob-Debug: %s not set\n", key)
	// 	fmt.Printf("Rob-Debug: Setting %s to true\n", key)
	// 	os.Setenv(key, "true")
	// } else {
	// 	fmt.Printf("Rob-Debug: %s=%s\n", key, val)
	// }

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	//Rob Added for Debug
	something, _ := wallet.List()
	log.Printf("Rob-Debug: wallet declared as: %v\n", something)

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	//Rob Added for Debug
	something, _ = wallet.List()
	log.Printf("Rob-Debug: wallet updated with appUser creds as: %v\n", something)

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	//Rob Added for Debug
	log.Printf("Rob-Debug: Gateway gw declared \n")

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	//Rob Added for Debug
	log.Printf("Rob-Debug: network declared as: %v\n", network.Name())

	contract := network.GetContract("basic")
	//Rob Added for Debug
	log.Printf("Rob-Debug: Contract Declared as: %v\n", contract.Name())

	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		cli, err1 := client.NewEnvClient()
		if err1 != nil {
			panic(err1)
		}
		containers, err1 := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err1 != nil {
			panic(err1)
		}
		for _, container := range containers {
			if strings.Contains(container.Image, "orderer") {
				fmt.Printf(" %-23s %v\n", "Name", "IMAGE")
				fmt.Printf(" %s  %v\n", container.Names, container.Image)
				fmt.Printf(" %-23s %v\n", "Name", "ID")
				fmt.Printf(" %s  %v\n", container.Names, container.ID)
				fmt.Printf(" %-23s %v\n", "Name", "Network Settings")
				fmt.Printf(" %s  %v\n", container.Names, container.NetworkSettings)
				fmt.Printf(" %-23s %v\n", "Name", "Ports")
				fmt.Printf(" %s  %v\n", container.Names, container.Ports)
			}
		}
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	log.Println("--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")
	result, err = contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))

	log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")
	result, err = contract.SubmitTransaction("CreateAsset", "asset13", "yellow", "5", "Tom", "1300")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	log.Println("--> Evaluate Transaction: ReadAsset, function returns an asset with a given assetID")
	result, err = contract.EvaluateTransaction("ReadAsset", "asset13")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v\n", err)
	}
	log.Println(string(result))

	log.Println("--> Evaluate Transaction: AssetExists, function returns 'true' if an asset with given assetID exist")
	result, err = contract.EvaluateTransaction("AssetExists", "asset1")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v\n", err)
	}
	log.Println(string(result))

	log.Println("--> Submit Transaction: TransferAsset asset1, transfer to new owner of Tom")
	_, err = contract.SubmitTransaction("TransferAsset", "asset1", "Tom")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}

	log.Println("--> Evaluate Transaction: ReadAsset, function returns 'asset1' attributes")
	result, err = contract.EvaluateTransaction("ReadAsset", "asset1")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))
	log.Println("============ application-golang ends ============")
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

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

	return wallet.Put("appUser", identity)
}
