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
	log.Printf("=== DETAILED Transaction Analysis for Account: %s ===", accountTxs.Account.String())
	log.Printf("Found %d recent transactions (last fetched: %s)",
		len(accountTxs.Transactions), accountTxs.LastFetched.Format(time.RFC3339))

	for i, tx := range accountTxs.Transactions {
		log.Printf("\n[%d] ==================== TRANSACTION DETAILS ====================", i+1)

		// Basic transaction info
		log.Printf("Signature: %s", tx.Signature)
		log.Printf("Slot: %d", tx.Slot)

		if tx.BlockTime != nil {
			timestamp := time.Unix(*tx.BlockTime, 0)
			log.Printf("Block Time: %s", timestamp.Format(time.RFC3339))
		}

		// Transaction Meta analysis (most detailed info here)
		if tx.Meta != nil {
			log.Printf("\n--- TRANSACTION META ---")
			log.Printf("Fee: %d lamports (%.9f SOL)", tx.Meta.Fee, float64(tx.Meta.Fee)/1e9)

			// Status and Error
			if tx.Meta.Err != nil {
				log.Printf("Status: FAILED - %v", tx.Meta.Err)
			} else {
				log.Printf("Status: SUCCESS")
			}

			// Account balance changes
			if len(tx.Meta.PreBalances) > 0 && len(tx.Meta.PostBalances) > 0 {
				log.Printf("\n--- BALANCE CHANGES ---")
				for j, preBalance := range tx.Meta.PreBalances {
					if j < len(tx.Meta.PostBalances) {
						postBalance := tx.Meta.PostBalances[j]
						change := int64(postBalance) - int64(preBalance)
						if change != 0 {
							log.Printf("Account[%d]: %d → %d lamports (change: %+d)",
								j, preBalance, postBalance, change)
						}
					}
				}
			}

			// Token balance changes
			if len(tx.Meta.PreTokenBalances) > 0 || len(tx.Meta.PostTokenBalances) > 0 {
				log.Printf("\n--- TOKEN BALANCE CHANGES ---")
				log.Printf("Pre-token balances: %d entries", len(tx.Meta.PreTokenBalances))
				log.Printf("Post-token balances: %d entries", len(tx.Meta.PostTokenBalances))

				// Show token balance details
				for _, tokenBalance := range tx.Meta.PostTokenBalances {
					if tokenBalance.UiTokenAmount != nil {
						log.Printf("Token: %s, Amount: %s",
							tokenBalance.Mint.String(), tokenBalance.UiTokenAmount.UiAmountString)
					}
				}
			}

			// Compute units consumed
			if tx.Meta.ComputeUnitsConsumed != nil {
				log.Printf("Compute Units Consumed: %d", *tx.Meta.ComputeUnitsConsumed)
			}

			// Log messages (program execution logs)
			if len(tx.Meta.LogMessages) > 0 {
				log.Printf("\n--- PROGRAM LOGS ---")
				for j, logMsg := range tx.Meta.LogMessages {
					if j < 5 { // Limit to first 5 logs to avoid spam
						log.Printf("Log[%d]: %s", j, logMsg)
					}
				}
				if len(tx.Meta.LogMessages) > 5 {
					log.Printf("... and %d more log messages", len(tx.Meta.LogMessages)-5)
				}
			}

			// Loaded addresses (for address lookup tables)
			if len(tx.Meta.LoadedAddresses.Writable) > 0 {
				log.Printf("Loaded Writable Addresses: %d", len(tx.Meta.LoadedAddresses.Writable))
			}
			if len(tx.Meta.LoadedAddresses.ReadOnly) > 0 {
				log.Printf("Loaded Readonly Addresses: %d", len(tx.Meta.LoadedAddresses.ReadOnly))
			}
		}

		// Transaction Message analysis
		if tx.Transaction != nil {
			msg := tx.Transaction.Message
			log.Printf("\n--- TRANSACTION MESSAGE ---")
			log.Printf("Recent Blockhash: %s", msg.RecentBlockhash.String())

			// Header info
			log.Printf("Required Signatures: %d", msg.Header.NumRequiredSignatures)
			log.Printf("Readonly Signed Accounts: %d", msg.Header.NumReadonlySignedAccounts)
			log.Printf("Readonly Unsigned Accounts: %d", msg.Header.NumReadonlyUnsignedAccounts)

			// Account keys
			log.Printf("Total Account Keys: %d", len(msg.AccountKeys))
			if len(msg.AccountKeys) > 0 {
				log.Printf("Account Keys:")
				for j, account := range msg.AccountKeys {
					if j < 5 { // Limit to first 5 accounts
						log.Printf("  [%d] %s", j, account.String())
					}
				}
				if len(msg.AccountKeys) > 5 {
					log.Printf("  ... and %d more accounts", len(msg.AccountKeys)-5)
				}
			}

			// Instructions analysis
			log.Printf("\n--- INSTRUCTIONS (%d total) ---", len(msg.Instructions))
			for j, instr := range msg.Instructions {
				if j < 3 { // Limit to first 3 instructions for readability
					log.Printf("Instruction[%d]:", j)
					log.Printf("  Program ID Index: %d", instr.ProgramIDIndex)
					if int(instr.ProgramIDIndex) < len(msg.AccountKeys) {
						log.Printf("  Program ID: %s", msg.AccountKeys[instr.ProgramIDIndex].String())
					}
					log.Printf("  Accounts: %v", instr.Accounts)
					log.Printf("  Data length: %d bytes", len(instr.Data))
				}
			}
			if len(msg.Instructions) > 3 {
				log.Printf("... and %d more instructions", len(msg.Instructions)-3)
			}

			// Address table lookups (for versioned transactions)
			if len(msg.AddressTableLookups) > 0 {
				log.Printf("\nAddress Table Lookups: %d", len(msg.AddressTableLookups))
			}
		}

		log.Printf("============================================================\n")
	}
}
