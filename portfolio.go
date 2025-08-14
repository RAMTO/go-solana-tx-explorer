package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// UserPortfolioService is responsible for fetching and displaying all SPL token
// holdings for a given wallet. It uses the JSON-RPC method `getTokenAccountsByOwner`
// with `jsonParsed` encoding to avoid manual binary decoding of token account data.
//
// We intentionally keep this file small and focused on a single responsibility.
type UserPortfolioService struct {
	client *rpc.Client
}

// NewUserPortfolioService creates a new portfolio service instance.
func NewUserPortfolioService(client *rpc.Client) *UserPortfolioService {
	return &UserPortfolioService{client: client}
}

// tokenProgramID is the well-known SPL Token Program ID (Tokenkeg...).
// Keeping it as a constant improves readability and avoids magic strings.
const tokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"

// PrintUserTokens fetches all SPL token accounts owned by `owner` and prints a
// concise portfolio table containing mint and balance details. Only non-zero
// balances are displayed to keep output relevant.
func (s *UserPortfolioService) PrintUserTokens(ctx context.Context, owner solana.PublicKey) error {
	// Raw JSON-RPC call (avoids mismatches in typed wrappers across versions)
	rpcURL := GetRPCURL()
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenAccountsByOwner",
		"params": []interface{}{
			owner.String(),
			map[string]interface{}{"programId": tokenProgramID},
			map[string]interface{}{"encoding": "jsonParsed", "commitment": "confirmed"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("rpc http request failed: %w", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read rpc response: %w", err)
	}

	var rpcResp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Value []struct {
				Account struct {
					Data struct {
						Parsed struct {
							Info struct {
								Mint        string `json:"mint"`
								TokenAmount struct {
									UiAmountString string `json:"uiAmountString"`
									Decimals       int    `json:"decimals"`
								} `json:"tokenAmount"`
							} `json:"info"`
						} `json:"parsed"`
					} `json:"data"`
				} `json:"account"`
			} `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return fmt.Errorf("decode rpc response: %w", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Load token registry for name/symbol enrichment (best-effort)
	registry, err := LoadDefaultRegistry(ctx)
	if err != nil {
		// Non-fatal; continue without enrichment
		registry = map[string]TokenInfo{}
	}

	// Collect holdings in a structured slice
	holdings := make([]TokenHolding, 0)
	for _, item := range rpcResp.Result.Value {
		mint := item.Account.Data.Parsed.Info.Mint
		amt := item.Account.Data.Parsed.Info.TokenAmount.UiAmountString
		decimals := item.Account.Data.Parsed.Info.TokenAmount.Decimals
		if amt == "" || amt == "0" || amt == "0.0" || amt == "0.00" || amt == "0.000" {
			continue
		}
		name := ""
		symbol := ""
		if info, ok := registry[mint]; ok {
			name = info.Name
			symbol = info.Symbol
		} else if mint == "So11111111111111111111111111111111111111112" {
			name = "Wrapped SOL"
			symbol = "wSOL"
		}
		holdings = append(holdings, TokenHolding{Mint: mint, UiAmount: amt, Decimals: decimals, Name: name, Symbol: symbol})
	}

	// Use the existing pretty formatter to display
	// Order by amount descending for better readability
	sort.Slice(holdings, func(i, j int) bool {
		ai, _ := strconv.ParseFloat(holdings[i].UiAmount, 64)
		aj, _ := strconv.ParseFloat(holdings[j].UiAmount, 64)
		return ai > aj
	})
	formatter := NewTransactionFormatter(false)
	formatter.FormatUserPortfolio(owner, holdings)
	return nil
}
