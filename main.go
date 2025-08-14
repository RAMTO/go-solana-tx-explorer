package main

import (
	"context"
	"log"

	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	rpcURL := GetRPCURL()
	client := rpc.New(rpcURL)
	ctx := context.Background()
	transactionService := NewTransactionService(client)
	portfolioService := NewUserPortfolioService(client)

	accountsToMonitor := GetWalletAddress()

	log.Println("Solana Transaction Monitor Starting...")

	account, err := GetAccountFromPublicKey(accountsToMonitor)
	if err != nil {
		log.Printf("Invalid account address %s: %v", accountsToMonitor, err)
	}

	accountTxs, err := transactionService.FetchAccountTransactions(ctx, account, TRANSACTIONS_LIMIT)
	if err != nil {
		log.Printf("Error fetching transactions for account %s: %v", account.String(), err)
	}

	if len(accountTxs.Transactions) > 0 {
		transactionService.AnalyzeTransactions(accountTxs)
	} else {
		log.Printf("No recent transactions found for account: %s", account.String())
	}

	if err := portfolioService.PrintUserTokens(ctx, account); err != nil {
		log.Printf("Error printing user tokens: %v", err)
	}
}
