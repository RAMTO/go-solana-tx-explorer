package main

import (
	"errors"
	"log"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

func GetAccountFromPublicKey(pubKey string) (solana.PublicKey, error) {
	account, err := solana.PublicKeyFromBase58(pubKey)
	if err != nil {
		return solana.PublicKey{}, errors.New("cannot get account from public key")
	}

	return account, nil
}

func GetRPCURL() string {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatal("RPC_URL environment variable is required")
	}
	return rpcURL
}

func GetWalletAddress() string {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	walletAddr := os.Getenv("WALLET_ADDRESS")
	if walletAddr == "" {
		log.Fatal("WALLET_ADDRESS environment variable is required")
	}
	return walletAddr
}
