package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// TokenInfo is a minimal entry from the Solana token list registry.
type TokenInfo struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
}

// tokenListResponse matches the root structure of the public token list.
type tokenListResponse struct {
	Tokens []TokenInfo `json:"tokens"`
}

var (
	registryOnce sync.Once
	registryData map[string]TokenInfo
	registryErr  error
)

// LoadDefaultRegistry merges multiple sources to maximize coverage.
// Order of precedence: Jupiter (all) > Jupiter (strict) > solana-labs list.
func LoadDefaultRegistry(ctx context.Context) (map[string]TokenInfo, error) {
	registryOnce.Do(func() {
		merged := make(map[string]TokenInfo)

		// 1) Jupiter ALL list (largest coverage)
		if m, err := loadJupiterList(ctx, "https://token.jup.ag/all"); err == nil {
			for k, v := range m {
				merged[k] = v
			}
		}
		// 2) Jupiter STRICT list
		if m, err := loadJupiterList(ctx, "https://token.jup.ag/strict"); err == nil {
			for k, v := range m {
				if _, ok := merged[k]; !ok {
					merged[k] = v
				}
			}
		}
		// 3) Legacy solana-labs list as last resort
		if m, err := loadSolanaLabsList(ctx); err == nil {
			for k, v := range m {
				if _, ok := merged[k]; !ok {
					merged[k] = v
				}
			}
		}

		if len(merged) == 0 {
			registryErr = fmt.Errorf("no token registry sources available")
			return
		}
		registryData = merged
	})
	return registryData, registryErr
}

func loadJupiterList(ctx context.Context, url string) (map[string]TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build jupiter req: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jupiter: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jupiter http status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read jupiter: %w", err)
	}

	var items []struct {
		Address string `json:"address"`
		Symbol  string `json:"symbol"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("decode jupiter: %w", err)
	}
	out := make(map[string]TokenInfo, len(items))
	for _, it := range items {
		if it.Address == "" {
			continue
		}
		out[it.Address] = TokenInfo{Address: it.Address, Symbol: it.Symbol, Name: it.Name}
	}
	return out, nil
}

func loadSolanaLabsList(ctx context.Context) (map[string]TokenInfo, error) {
	const tokenListURL = "https://cdn.jsdelivr.net/gh/solana-labs/token-list@main/src/tokens/solana.tokenlist.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build registry request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch registry: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry http status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}
	var data tokenListResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode registry: %w", err)
	}
	byMint := make(map[string]TokenInfo, len(data.Tokens))
	for _, t := range data.Tokens {
		byMint[t.Address] = t
	}
	return byMint, nil
}
