package main

import (
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// TransactionFormatter handles pretty printing of transaction data
type TransactionFormatter struct {
	showFullData bool
}

// NewTransactionFormatter creates a new formatter instance
func NewTransactionFormatter(showFullData bool) *TransactionFormatter {
	return &TransactionFormatter{
		showFullData: showFullData,
	}
}

// FormatTransactionSummary displays a summary table of all transactions
func (f *TransactionFormatter) FormatTransactionSummary(accountTxs *AccountTransactions) {
	// Print header with account info
	fmt.Printf("\n%s\n", text.Colors{text.BgBlue, text.FgWhite}.Sprint(" SOLANA TRANSACTION EXPLORER "))
	fmt.Printf("Account: %s\n", text.FgCyan.Sprint(accountTxs.Account.String()))
	fmt.Printf("Total Transactions: %s\n", text.FgGreen.Sprint(len(accountTxs.Transactions)))
	fmt.Printf("Last Fetched: %s\n\n", text.FgYellow.Sprint(accountTxs.LastFetched.Format(time.RFC3339)))

	// Create summary table
	t := table.NewWriter()
	t.SetTitle("Transaction Summary")
	t.AppendHeader(table.Row{"#", "Signature (Short)", "Status", "Slot", "Time", "Fee (SOL)", "Balance Change"})

	for i, tx := range accountTxs.Transactions {
		// Truncate signature for readability
		shortSig := tx.Signature
		if len(shortSig) > 16 {
			shortSig = shortSig[:8] + "..." + shortSig[len(shortSig)-8:]
		}

		// Determine status
		status := "✅ SUCCESS"
		if tx.Meta != nil && tx.Meta.Err != nil {
			status = "❌ FAILED"
		}

		// Format time
		timeStr := "N/A"
		if tx.BlockTime != nil {
			timestamp := time.Unix(*tx.BlockTime, 0)
			timeStr = timestamp.Format("01-02 15:04")
		}

		// Calculate fee in SOL
		feeSOL := "0"
		if tx.Meta != nil {
			feeSOL = fmt.Sprintf("%.6f", float64(tx.Meta.Fee)/1e9)
		}

		// Calculate balance change for the main account
		balanceChange := "0"
		if tx.Meta != nil && len(tx.Meta.PreBalances) > 0 && len(tx.Meta.PostBalances) > 0 {
			change := int64(tx.Meta.PostBalances[0]) - int64(tx.Meta.PreBalances[0])
			if change != 0 {
				balanceChange = fmt.Sprintf("%+.6f", float64(change)/1e9)
			}
		}

		t.AppendRow(table.Row{
			i + 1,
			shortSig,
			status,
			tx.Slot,
			timeStr,
			feeSOL,
			balanceChange,
		})
	}

	// Style the table
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateRows = true

	fmt.Println(t.Render())
}

// FormatTransactionDetails displays detailed information for a specific transaction
func (f *TransactionFormatter) FormatTransactionDetails(tx TransactionInfo, index int) {
	fmt.Printf("\n%s\n",
		text.Colors{text.BgGreen, text.FgWhite}.Sprintf(" TRANSACTION #%d DETAILS ", index+1))

	// Basic info table
	basicInfo := table.NewWriter()
	basicInfo.SetTitle("Basic Information")
	basicInfo.AppendRow(table.Row{"Signature", tx.Signature})
	basicInfo.AppendRow(table.Row{"Slot", tx.Slot})

	if tx.BlockTime != nil {
		timestamp := time.Unix(*tx.BlockTime, 0)
		basicInfo.AppendRow(table.Row{"Block Time", timestamp.Format(time.RFC3339)})
	}

	basicInfo.SetStyle(table.StyleColoredDark)
	fmt.Println(basicInfo.Render())

	// Transaction meta information
	if tx.Meta != nil {
		f.formatTransactionMeta(tx.Meta)
	}

	// Transaction message information
	if tx.Transaction != nil {
		f.formatTransactionMessage(tx.Transaction)
	}
}

// formatTransactionMeta formats the transaction metadata
func (f *TransactionFormatter) formatTransactionMeta(meta *rpc.TransactionMeta) {
	fmt.Printf("\n%s\n", text.FgYellow.Sprint("💰 TRANSACTION META"))

	metaTable := table.NewWriter()
	metaTable.SetTitle("Meta Information")

	// Fee
	metaTable.AppendRow(table.Row{"Fee (lamports)", fmt.Sprintf("%d", meta.Fee)})
	metaTable.AppendRow(table.Row{"Fee (SOL)", fmt.Sprintf("%.9f", float64(meta.Fee)/1e9)})

	// Status
	status := "SUCCESS ✅"
	if meta.Err != nil {
		status = fmt.Sprintf("FAILED ❌ - %v", meta.Err)
	}
	metaTable.AppendRow(table.Row{"Status", status})

	// Compute units
	if meta.ComputeUnitsConsumed != nil {
		metaTable.AppendRow(table.Row{"Compute Units", fmt.Sprintf("%d", *meta.ComputeUnitsConsumed)})
	}

	metaTable.SetStyle(table.StyleLight)
	fmt.Println(metaTable.Render())

	// Balance changes
	f.formatBalanceChanges(meta)

	// Token balance changes
	f.formatTokenBalances(meta)

	// Program logs (limited)
	if len(meta.LogMessages) > 0 {
		f.formatProgramLogs(meta.LogMessages)
	}
}

// formatBalanceChanges displays SOL balance changes
func (f *TransactionFormatter) formatBalanceChanges(meta *rpc.TransactionMeta) {
	if len(meta.PreBalances) == 0 || len(meta.PostBalances) == 0 {
		return
	}

	fmt.Printf("\n%s\n", text.FgCyan.Sprint("📊 SOL BALANCE CHANGES"))

	balanceTable := table.NewWriter()
	balanceTable.SetTitle("Account Balance Changes")
	balanceTable.AppendHeader(table.Row{"Account", "Pre (SOL)", "Post (SOL)", "Change (SOL)"})

	for i, preBalance := range meta.PreBalances {
		if i < len(meta.PostBalances) {
			postBalance := meta.PostBalances[i]
			change := int64(postBalance) - int64(preBalance)

			if change != 0 {
				changeStr := fmt.Sprintf("%+.6f", float64(change)/1e9)
				if change > 0 {
					changeStr = text.FgGreen.Sprint(changeStr)
				} else {
					changeStr = text.FgRed.Sprint(changeStr)
				}

				balanceTable.AppendRow(table.Row{
					fmt.Sprintf("Account[%d]", i),
					fmt.Sprintf("%.6f", float64(preBalance)/1e9),
					fmt.Sprintf("%.6f", float64(postBalance)/1e9),
					changeStr,
				})
			}
		}
	}

	if balanceTable.Length() > 0 {
		balanceTable.SetStyle(table.StyleLight)
		fmt.Println(balanceTable.Render())
	}
}

// formatTokenBalances displays token balance information
func (f *TransactionFormatter) formatTokenBalances(meta *rpc.TransactionMeta) {
	if len(meta.PostTokenBalances) == 0 {
		return
	}

	fmt.Printf("\n%s\n", text.FgMagenta.Sprint("🪙 TOKEN BALANCES"))

	tokenTable := table.NewWriter()
	tokenTable.SetTitle("Token Information")
	tokenTable.AppendHeader(table.Row{"Mint", "Amount", "Decimals"})

	for _, tokenBalance := range meta.PostTokenBalances {
		if tokenBalance.UiTokenAmount != nil {
			tokenTable.AppendRow(table.Row{
				tokenBalance.Mint.String()[:8] + "...",
				tokenBalance.UiTokenAmount.UiAmountString,
				tokenBalance.UiTokenAmount.Decimals,
			})
		}
	}

	tokenTable.SetStyle(table.StyleLight)
	fmt.Println(tokenTable.Render())
}

// formatProgramLogs displays program execution logs
func (f *TransactionFormatter) formatProgramLogs(logs []string) {
	fmt.Printf("\n%s\n", text.FgYellow.Sprint("📝 PROGRAM LOGS"))

	logTable := table.NewWriter()
	logTable.SetTitle("Program Execution Logs")
	logTable.AppendHeader(table.Row{"#", "Message"})

	maxLogs := 5
	if f.showFullData {
		maxLogs = len(logs)
	}

	for i, logMsg := range logs {
		if i >= maxLogs {
			break
		}

		// Truncate very long log messages
		if len(logMsg) > 80 && !f.showFullData {
			logMsg = logMsg[:77] + "..."
		}

		logTable.AppendRow(table.Row{i + 1, logMsg})
	}

	if len(logs) > maxLogs {
		logTable.AppendRow(table.Row{"...", fmt.Sprintf("and %d more logs", len(logs)-maxLogs)})
	}

	logTable.SetStyle(table.StyleLight)
	logTable.Style().Options.SeparateRows = true
	fmt.Println(logTable.Render())
}

// formatTransactionMessage displays transaction message details
func (f *TransactionFormatter) formatTransactionMessage(tx *solana.Transaction) {
	fmt.Printf("\n%s\n", text.FgBlue.Sprint("📄 TRANSACTION MESSAGE"))

	msg := tx.Message

	// Basic message info
	msgTable := table.NewWriter()
	msgTable.SetTitle("Message Information")
	msgTable.AppendRow(table.Row{"Recent Blockhash", msg.RecentBlockhash.String()})
	msgTable.AppendRow(table.Row{"Required Signatures", msg.Header.NumRequiredSignatures})
	msgTable.AppendRow(table.Row{"Readonly Signed", msg.Header.NumReadonlySignedAccounts})
	msgTable.AppendRow(table.Row{"Readonly Unsigned", msg.Header.NumReadonlyUnsignedAccounts})
	msgTable.AppendRow(table.Row{"Total Accounts", len(msg.AccountKeys)})
	msgTable.AppendRow(table.Row{"Total Instructions", len(msg.Instructions)})

	msgTable.SetStyle(table.StyleLight)
	fmt.Println(msgTable.Render())

	// Account keys
	if len(msg.AccountKeys) > 0 {
		f.formatAccountKeys(msg.AccountKeys)
	}

	// Instructions
	if len(msg.Instructions) > 0 {
		f.formatInstructions(msg.Instructions, msg.AccountKeys)
	}
}

// formatAccountKeys displays account keys used in the transaction
func (f *TransactionFormatter) formatAccountKeys(accountKeys []solana.PublicKey) {
	fmt.Printf("\n%s\n", text.FgGreen.Sprint("🔑 ACCOUNT KEYS"))

	accountTable := table.NewWriter()
	accountTable.SetTitle("Transaction Account Keys")
	accountTable.AppendHeader(table.Row{"Index", "Public Key"})

	maxAccounts := 5
	if f.showFullData {
		maxAccounts = len(accountKeys)
	}

	for i, account := range accountKeys {
		if i >= maxAccounts {
			break
		}
		accountTable.AppendRow(table.Row{i, account.String()})
	}

	if len(accountKeys) > maxAccounts {
		accountTable.AppendRow(table.Row{"...", fmt.Sprintf("and %d more accounts", len(accountKeys)-maxAccounts)})
	}

	accountTable.SetStyle(table.StyleLight)
	fmt.Println(accountTable.Render())
}

// formatInstructions displays transaction instructions
func (f *TransactionFormatter) formatInstructions(instructions []solana.CompiledInstruction, accountKeys []solana.PublicKey) {
	fmt.Printf("\n%s\n", text.FgRed.Sprint("⚙️ INSTRUCTIONS"))

	instrTable := table.NewWriter()
	instrTable.SetTitle("Transaction Instructions")
	instrTable.AppendHeader(table.Row{"#", "Program", "Accounts", "Data Size"})

	maxInstr := 3
	if f.showFullData {
		maxInstr = len(instructions)
	}

	for i, instr := range instructions {
		if i >= maxInstr {
			break
		}

		programID := "Unknown"
		if int(instr.ProgramIDIndex) < len(accountKeys) {
			programID = accountKeys[instr.ProgramIDIndex].String()[:8] + "..."
		}

		accounts := fmt.Sprintf("%v", instr.Accounts)
		if len(accounts) > 20 {
			accounts = accounts[:17] + "..."
		}

		instrTable.AppendRow(table.Row{
			i + 1,
			programID,
			accounts,
			fmt.Sprintf("%d bytes", len(instr.Data)),
		})
	}

	if len(instructions) > maxInstr {
		instrTable.AppendRow(table.Row{"...", fmt.Sprintf("and %d more instructions", len(instructions)-maxInstr), "", ""})
	}

	instrTable.SetStyle(table.StyleLight)
	fmt.Println(instrTable.Render())
}
