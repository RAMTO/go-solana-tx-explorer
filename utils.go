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

// GetWSURL returns the WebSocket RPC URL. If WS_URL is not set, it tries to
// derive it from RPC_URL by replacing the scheme with wss:// when possible.
func GetWSURL() string {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	wsURL := os.Getenv("WS_URL")
	if wsURL != "" {
		return wsURL
	}
	httpURL := os.Getenv("RPC_URL")
	if httpURL == "" {
		log.Fatal("WS_URL or RPC_URL environment variable is required")
	}
	// naive derive: support https:// → wss://, http:// → ws://
	if len(httpURL) >= 8 && httpURL[:8] == "https://" {
		return "wss://" + httpURL[8:]
	}
	if len(httpURL) >= 7 && httpURL[:7] == "http://" {
		return "ws://" + httpURL[7:]
	}
	// if already ws/wss, return as-is
	if len(httpURL) >= 6 && httpURL[:6] == "wss://" {
		return httpURL
	}
	if len(httpURL) >= 5 && httpURL[:5] == "ws://" {
		return httpURL
	}
	return httpURL
}
