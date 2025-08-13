# Complete List of Extractable Solana Transaction Data

This document details all the data fields that can be extracted from a Solana transaction using the `solana-go` library.

## Currently Extracted (Basic Analysis)

✅ **Transaction Signature** - Unique transaction identifier  
✅ **Slot** - Block slot number  
✅ **Block Time** - Transaction timestamp  
✅ **Fee** - Transaction fee in lamports  
✅ **Status** - Success/Error status  
✅ **Instructions Count** - Number of instructions in transaction

## Additional Extractable Fields

### 📊 Transaction Meta (`tx.Meta`)

#### Balance Changes

- **PreBalances** - Account balances before transaction (array of uint64)
- **PostBalances** - Account balances after transaction (array of uint64)
- **Balance Differences** - Calculated change per account

#### Token Information

- **PreTokenBalances** - Token balances before transaction
- **PostTokenBalances** - Token balances after transaction
- **Token Mint Addresses** - SPL token contract addresses
- **Token UI Amounts** - Human-readable token amounts

#### Execution Details

- **ComputeUnitsConsumed** - Compute units used by transaction
- **LogMessages** - Program execution logs and debug info
- **InnerInstructions** - Nested instruction calls
- **LoadedAddresses** - Addresses loaded via lookup tables
  - Writable addresses
  - Readonly addresses

#### Return Data

- **ReturnData** - Data returned by programs
- **ReturnDataProgram** - Program that returned data

### 🔄 Transaction Message (`tx.Transaction.Message`)

#### Header Information

- **NumRequiredSignatures** - Number of required signatures
- **NumReadonlySignedAccounts** - Readonly signed accounts count
- **NumReadonlyUnsignedAccounts** - Readonly unsigned accounts count

#### Account Information

- **AccountKeys** - All public keys involved in transaction
- **RecentBlockhash** - Recent blockhash for transaction validity

#### Instructions Details

For each instruction:

- **ProgramIDIndex** - Index of the program in account keys
- **Program ID** - Actual program public key
- **Accounts** - Array of account indices used by instruction
- **Data** - Instruction data (program-specific)

#### Advanced Features

- **AddressTableLookups** - Address lookup table references (for versioned transactions)
- **Signatures** - All transaction signatures

### 💰 Financial Analysis Possibilities

#### SOL Transfers

- Calculate SOL transfers between accounts
- Identify sender and receiver addresses
- Track balance changes in SOL equivalent

#### Token Transfers

- SPL token transfers and amounts
- Token mint identification
- Multi-token transaction analysis

#### Fee Analysis

- Fee calculation and breakdown
- Fee per compute unit analysis
- Transaction cost efficiency

### 🔍 Advanced Analytics

#### Program Interaction

- Identify which programs were called
- Analyze program instruction patterns
- Track cross-program invocations

#### Address Analysis

- Identify unique addresses involved
- Track address interaction patterns
- Account role classification (signer, writable, readonly)

#### Transaction Complexity

- Instruction count and types
- Compute unit consumption analysis
- Transaction size and complexity metrics

## Implementation Examples

### Balance Change Analysis

```go
// Calculate SOL balance changes
for i, preBalance := range tx.Meta.PreBalances {
    if i < len(tx.Meta.PostBalances) {
        change := int64(tx.Meta.PostBalances[i]) - int64(preBalance)
        solChange := float64(change) / 1e9 // Convert to SOL
        fmt.Printf("Account[%d]: %+.9f SOL", i, solChange)
    }
}
```

### Token Transfer Detection

```go
// Detect token transfers
for _, tokenBalance := range tx.Meta.PostTokenBalances {
    if tokenBalance.Mint != nil {
        fmt.Printf("Token: %s, Amount: %s",
            tokenBalance.Mint.String(),
            tokenBalance.UiTokenAmount.UiAmountString)
    }
}
```

### Program Identification

```go
// Identify programs used
for _, instr := range tx.Transaction.Message.Instructions {
    if int(instr.ProgramIDIndex) < len(tx.Transaction.Message.AccountKeys) {
        programID := tx.Transaction.Message.AccountKeys[instr.ProgramIDIndex]
        fmt.Printf("Program: %s", programID.String())
    }
}
```

## Useful Program IDs to Recognize

- **System Program**: `11111111111111111111111111111112`
- **SPL Token Program**: `TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA`
- **SPL Associated Token**: `ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL`
- **Serum DEX**: `9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin`
- **Raydium**: `675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8`

## Performance Considerations

- **Log Messages** can be extensive - limit display for readability
- **Instructions** may be numerous - consider pagination
- **Account Keys** can be large - implement smart filtering
- **Token Balances** require additional processing - cache mint info

## Enhanced Analysis Functions

The codebase now includes:

- `AnalyzeTransactions()` - Basic analysis (current)
- `AnalyzeTransactionsDetailed()` - Comprehensive analysis (new)

Use the detailed analysis for deep investigation and the basic analysis for quick overviews.
