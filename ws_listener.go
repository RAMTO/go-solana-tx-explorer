package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// ListenWalletTransactions provides a minimal "live" listener using HTTP polling
// as a fallback to WebSocket streaming. We derive an HTTP RPC URL from the given
// wsURL and repeatedly call getSignaturesForAddress, printing any new signatures.
// This keeps dependencies minimal and works against Helius endpoints too.
func ListenWalletTransactions(ctx context.Context, wsURL string, wallet solana.PublicKey) error {
	httpURL := wsURL
	if strings.HasPrefix(httpURL, "wss://") {
		httpURL = "https://" + strings.TrimPrefix(httpURL, "wss://")
	} else if strings.HasPrefix(httpURL, "ws://") {
		httpURL = "http://" + strings.TrimPrefix(httpURL, "ws://")
	}

	client := rpc.New(httpURL)
	log.Printf("🔌 Listening (poll) for transactions mentioning %s ...", wallet.String())

	seen := make(map[string]struct{})
	// Seed with current known signatures so we only report NEW ones going forward
	if sigs, err := client.GetSignaturesForAddress(ctx, wallet); err == nil {
		for _, s := range sigs {
			seen[s.Signature.String()] = struct{}{}
		}
	}
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sigs, err := client.GetSignaturesForAddress(ctx, wallet)
			if err != nil {
				log.Printf("poll error: %v", err)
				continue
			}
			// Iterate in reverse so older new entries are printed first
			for i := len(sigs) - 1; i >= 0; i-- {
				s := sigs[i]
				sigStr := s.Signature.String()
				if _, ok := seen[sigStr]; ok {
					continue
				}
				seen[sigStr] = struct{}{}
				log.Printf("🆕 Tx observed: %s (slot %d)", sigStr, s.Slot)
			}
		}
	}
}
