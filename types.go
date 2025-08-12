package main

import (
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransactionInfo struct {
	Signature   string               `json:"signature"`
	Slot        uint64               `json:"slot"`
	BlockTime   *int64               `json:"blockTime,omitempty"`
	Meta        *rpc.TransactionMeta `json:"meta,omitempty"`
	Transaction *solana.Transaction  `json:"transaction,omitempty"`
}

type AccountTransactions struct {
	Account      solana.PublicKey  `json:"account"`
	Transactions []TransactionInfo `json:"transactions"`
	LastFetched  time.Time         `json:"last_fetched"`
}
