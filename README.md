# Solana Transaction Explorer

A Go-based command-line tool for monitoring and analyzing Solana blockchain transactions for specific wallet addresses.

## Features

- **Real-time Transaction Monitoring**: Fetches and displays recent transactions for any Solana wallet address
- **Beautiful Console Visualization**: Pretty-formatted tables with colors and structured layouts using go-pretty
- **Detailed Transaction Analysis**: Shows transaction signatures, slots, timestamps, fees, status, and instruction counts
- **Environment-based Configuration**: Securely manages RPC URLs and wallet addresses through environment variables
- **Interactive Summary Tables**: Overview table showing all transactions with key metrics at a glance
- **Detailed Transaction Views**: In-depth analysis of individual transactions with formatted sections

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

- `TRANSACTIONS_LIMIT`: Number of recent transactions to fetch (default: 5)

### New Console Visualization Features

The application now uses the **go-pretty** library to provide beautiful console output:

- **📊 Transaction Summary Table**: Overview of all transactions with status, fees, and balance changes
- **💰 Transaction Meta Information**: Detailed fee analysis, compute units, and status
- **📋 Balance Changes**: SOL balance changes with color-coded positive/negative values
- **🪙 Token Information**: Token balance details and mint addresses
- **📝 Program Logs**: Execution logs from Solana programs
- **🔑 Account Keys**: Transaction account participants
- **⚙️ Instructions**: Program instructions with data sizes

### Output Customization

You can customize the output detail level by modifying the `TransactionFormatter`:

```go
// For detailed output (shows all data)
formatter := NewTransactionFormatter(true)

// For concise output (limited data)
formatter := NewTransactionFormatter(false)
```

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
