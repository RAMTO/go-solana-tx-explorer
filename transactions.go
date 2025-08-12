package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransactionService struct {
	client *rpc.Client
}

func NewTransactionService(client *rpc.Client) *TransactionService {
	return &TransactionService{client}
}

func (t *TransactionService) FetchAccountTransactions(ctx context.Context, account solana.PublicKey, limit int) (*AccountTransactions, error) {
	signatures, err := t.client.GetSignaturesForAddress(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures for account %s: %w", account.String(), err)
	}

	processCount := len(signatures)
	if limit > 0 && limit < processCount {
		processCount = limit
	}

	fmt.Printf("Count: %d \n", processCount)

	type transactionResult struct {
		info  TransactionInfo
		index int
		err   error
	}

	resultChan := make(chan transactionResult, processCount)
	var wg sync.WaitGroup

	for i := 0; i < processCount; i++ {
		wg.Add(1)
		go func(index int, sig *rpc.TransactionSignature) {
			defer wg.Done()

			maxVersion := uint64(0)
			txResult, err := t.client.GetTransaction(ctx, sig.Signature, &rpc.GetTransactionOpts{
				Encoding:                       solana.EncodingBase64,
				Commitment:                     rpc.CommitmentConfirmed,
				MaxSupportedTransactionVersion: &maxVersion,
			})

			if err != nil {
				log.Printf("Failed to get transaction %s: %v", sig.Signature.String(), err)
				resultChan <- transactionResult{err: err, index: index}
				return
			}

			var blockTime *int64
			if sig.BlockTime != nil {
				timestamp := int64(*sig.BlockTime)
				blockTime = &timestamp
			}

			txInfo := TransactionInfo{
				Signature: sig.Signature.String(),
				Slot:      sig.Slot,
				BlockTime: blockTime,
				Meta:      txResult.Meta,
			}

			if txResult.Transaction != nil {
				parsedTx, err := txResult.Transaction.GetTransaction()
				if err != nil {
					log.Printf("Failed to parse transaction %s: %v (will continue)", sig.Signature.String(), err)
				} else {
					txInfo.Transaction = parsedTx
				}
			}

			resultChan <- transactionResult{info: txInfo, index: index, err: nil}
		}(i, signatures[i])
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]TransactionInfo, 0, processCount)
	resultMap := make(map[int]TransactionInfo)

	for result := range resultChan {
		if result.err == nil {
			resultMap[result.index] = result.info
		}
	}

	for i := 0; i < processCount; i++ {
		if txInfo, exists := resultMap[i]; exists {
			results = append(results, txInfo)
		}
	}

	transactions := results

	return &AccountTransactions{
		Account:      account,
		Transactions: transactions,
		LastFetched:  time.Now(),
	}, nil
}

func (t *TransactionService) AnalyzeTransactions(accountTxs *AccountTransactions) {
	log.Printf("=== Transaction Analysis for Account: %s ===", accountTxs.Account.String())
	log.Printf("Found %d recent transactions (last fetched: %s)",
		len(accountTxs.Transactions), accountTxs.LastFetched.Format(time.RFC3339))

	for i, tx := range accountTxs.Transactions {
		log.Printf("  [%d] Signature: %s", i+1, tx.Signature)
		log.Printf("      Slot: %d", tx.Slot)

		if tx.BlockTime != nil {
			timestamp := time.Unix(*tx.BlockTime, 0)
			log.Printf("      Time: %s", timestamp.Format(time.RFC3339))
		}

		if tx.Meta != nil {
			log.Printf("      Fee: %d lamports", tx.Meta.Fee)
			if tx.Meta.Err != nil {
				log.Printf("      ERROR: %v", tx.Meta.Err)
			} else {
				log.Printf("      Status: Success")
			}
		}

		if tx.Transaction != nil && len(tx.Transaction.Message.Instructions) > 0 {
			log.Printf("      Instructions: %d", len(tx.Transaction.Message.Instructions))
		}
		log.Println()
	}
}
