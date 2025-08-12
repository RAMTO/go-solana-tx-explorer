# Solana Transaction Explorer

A Go-based command-line tool for monitoring and analyzing Solana blockchain transactions for specific wallet addresses.

## Features

- **Real-time Transaction Monitoring**: Fetches and displays recent transactions for any Solana wallet address
- **Detailed Transaction Analysis**: Shows transaction signatures, slots, timestamps, fees, status, and instruction counts
- **Environment-based Configuration**: Securely manages RPC URLs and wallet addresses through environment variables
- **Clean Output Format**: Well-formatted transaction logs with timestamps and structured data

## Prerequisites

- Go 1.24.5 or higher
- A Solana RPC endpoint (Helius, QuickNode, or any other provider)
- A valid Solana wallet address to monitor

## Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd go-solana-tx-explorer
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Set up environment variables:**
   Create a `.env` file in the project root:

   ```bash
   # Solana RPC Configuration
   RPC_URL=https://mainnet.helius-rpc.com/?api-key=YOUR_API_KEY

   # Wallet address to monitor
   WALLET_ADDRESS=YOUR_SOLANA_WALLET_ADDRESS
   ```

## Configuration

### Constants

- `TRANSACTIONS_LIMIT`: Number of recent transactions to fetch (default: 20)

## Usage

### Basic Usage

Run the transaction explorer:

```bash
go run .
```

### Build and Run

Build the executable:

```bash
go build -o solana-tx-explorer
./solana-tx-explorer
```
