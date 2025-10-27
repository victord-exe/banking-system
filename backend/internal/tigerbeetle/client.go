package tigerbeetle

import (
	"fmt"
	"log"
	"math/big"
	"net"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Client wraps the TigerBeetle client
type Client struct {
	client          tb.Client
	clusterID       uint128
	SystemAccountID uint128 // Bank system account for deposits/withdrawals
}

// uint128 represents a 128-bit unsigned integer
type uint128 = tb_types.Uint128

// NewClient creates a new TigerBeetle client
func NewClient(address string) (*Client, error) {
	log.Printf("🔗 Attempting to connect to TigerBeetle at: %s", address)

	// Resolve hostname to IP address (required for TigerBeetle Go client)
	resolvedAddr, err := resolveAddress(address)
	if err != nil {
		log.Printf("⚠️  Failed to resolve TigerBeetle address '%s': %v", address, err)
		log.Printf("⚠️  This might cause connection issues. Ensure TigerBeetle is running and accessible.")
		// Don't fail immediately, let the TigerBeetle client try with the original address
		resolvedAddr = address
	}

	log.Printf("🔍 TigerBeetle connection details:")
	log.Printf("   Original address: %s", address)
	log.Printf("   Resolved address: %s", resolvedAddr)

	// Cluster ID (must match the initialized cluster)
	clusterID := tb_types.ToUint128(0)

	// Create TigerBeetle client
	log.Printf("📡 Creating TigerBeetle client connection...")
	tbClient, err := tb.NewClient(clusterID, []string{resolvedAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to create TigerBeetle client to '%s' (resolved: '%s'): %w", address, resolvedAddr, err)
	}

	log.Printf("✅ TigerBeetle client connected successfully!")

	client := &Client{
		client:          tbClient,
		clusterID:       clusterID,
		SystemAccountID: tb_types.ToUint128(1), // System account ID = 1
	}

	// Initialize system account if needed
	log.Printf("🔧 Initializing system account...")
	if err := client.ensureSystemAccount(); err != nil {
		return nil, fmt.Errorf("failed to initialize system account: %w", err)
	}

	log.Printf("✅ TigerBeetle client fully initialized!")
	return client, nil
}

// ensureSystemAccount creates the system bank account if it doesn't exist
func (c *Client) ensureSystemAccount() error {
	// Try to create system account (ID = 1)
	accounts := []tb_types.Account{
		{
			ID:     c.SystemAccountID,
			Ledger: 1,   // User ledger
			Code:   999, // System account code
			Flags:  0,   // No restrictions on system account
		},
	}

	results, err := c.client.CreateAccounts(accounts)
	if err != nil {
		return fmt.Errorf("failed to create system account: %w", err)
	}

	// Check results - ignore "exists" error
	for _, result := range results {
		if result.Result != tb_types.AccountExists {
			log.Println("✅ System account initialized (ID: 1)")
		} else {
			log.Println("ℹ️  System account already exists")
		}
	}

	return nil
}

// CreateAccount creates a new user account in TigerBeetle
func (c *Client) CreateAccount(accountID uint64) error {
	id := tb_types.ToUint128(accountID)

	accounts := []tb_types.Account{
		{
			ID:     id,
			Ledger: 1, // User ledger
			Code:   1, // User account code
			Flags:  tb_types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
		},
	}

	results, err := c.client.CreateAccounts(accounts)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Check for errors in results
	if len(results) > 0 {
		return fmt.Errorf("failed to create account: result code %d", results[0].Result)
	}

	log.Printf("✅ Created TigerBeetle account ID: %d", accountID)

	return nil
}

// GetBalance retrieves the balance of an account
func (c *Client) GetBalance(accountID uint64) (int64, error) {
	log.Printf("🟢 [TigerBeetle] GetBalance called for accountID: %d", accountID)

	id := tb_types.ToUint128(accountID)
	log.Printf("🟢 [TigerBeetle] Converted to Uint128: %v", id)

	log.Printf("🟢 [TigerBeetle] Calling LookupAccounts...")
	accounts, err := c.client.LookupAccounts([]tb_types.Uint128{id})
	if err != nil {
		log.Printf("❌ [TigerBeetle] Failed to lookup account %d: %v", accountID, err)
		return 0, fmt.Errorf("failed to lookup account: %w", err)
	}
	log.Printf("🟢 [TigerBeetle] LookupAccounts returned %d account(s)", len(accounts))

	if len(accounts) == 0 {
		log.Printf("❌ [TigerBeetle] Account %d not found in TigerBeetle", accountID)
		return 0, fmt.Errorf("account not found")
	}

	account := accounts[0]
	log.Printf("🟢 [TigerBeetle] Account details: ID=%v, Ledger=%d, Code=%d", account.ID, account.Ledger, account.Code)
	log.Printf("🟢 [TigerBeetle] Account raw values: CreditsPosted=%v, DebitsPosted=%v", account.CreditsPosted, account.DebitsPosted)

	// Balance = Credits - Debits (for accounts with DebitsMustNotExceedCredits flag)
	// Convert Uint128 to big.Int for arithmetic
	creditsBI := account.CreditsPosted.BigInt()
	debitsBI := account.DebitsPosted.BigInt()
	log.Printf("🟢 [TigerBeetle] Credits as BigInt: %s", creditsBI.String())
	log.Printf("🟢 [TigerBeetle] Debits as BigInt: %s", debitsBI.String())

	balanceBI := new(big.Int).Sub(&creditsBI, &debitsBI)
	log.Printf("🟢 [TigerBeetle] Balance as BigInt: %s", balanceBI.String())

	// Convert to int64 (safe for reasonable banking amounts)
	balance := balanceBI.Int64()
	log.Printf("✅ [TigerBeetle] Final balance: %d cents (accountID: %d)", balance, accountID)

	return balance, nil
}

// CreateTransfers creates one or more transfers in TigerBeetle
func (c *Client) CreateTransfers(transfers []tb_types.Transfer) ([]tb_types.TransferEventResult, error) {
	results, err := c.client.CreateTransfers(transfers)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfers: %w", err)
	}
	return results, nil
}

// LookupAccounts retrieves account information from TigerBeetle
func (c *Client) LookupAccounts(accountIDs []tb_types.Uint128) ([]tb_types.Account, error) {
	accounts, err := c.client.LookupAccounts(accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup accounts: %w", err)
	}
	return accounts, nil
}

// resolveAddress resolves a hostname:port to IP:port for TigerBeetle client
func resolveAddress(address string) (string, error) {
	// Split address into host and port
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// If no port specified, assume it's just a hostname
		host = address
		port = "3000" // Default TigerBeetle port
	}

	// Check if host is already an IP address
	if net.ParseIP(host) != nil {
		// Already an IP, return as-is
		result := net.JoinHostPort(host, port)
		log.Printf("✅ Address is already an IP: %s", result)
		return result, nil
	}

	// Try to resolve hostname to IP address with retries
	log.Printf("🔍 Resolving hostname '%s' to IP address...", host)
	
	var ips []string
	var lastErr error
	
	// Retry DNS resolution up to 3 times with small delays
	for attempt := 1; attempt <= 3; attempt++ {
		ips, lastErr = net.LookupHost(host)
		if lastErr == nil && len(ips) > 0 {
			break
		}
		
		if attempt < 3 {
			log.Printf("⚠️  DNS resolution attempt %d/3 failed for '%s': %v (retrying...)", attempt, host, lastErr)
			// Simple sleep could be added here if needed
		}
	}
	
	if lastErr != nil {
		return "", fmt.Errorf("failed to lookup host '%s' after 3 attempts: %w", host, lastErr)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for host '%s'", host)
	}

	// Use the first resolved IP address
	resolvedIP := ips[0]
	resolvedAddr := net.JoinHostPort(resolvedIP, port)

	log.Printf("✅ DNS resolved '%s' → '%s' (from %d available IPs)", address, resolvedAddr, len(ips))

	return resolvedAddr, nil
}

// Close closes the TigerBeetle client connection
func (c *Client) Close() {
	c.client.Close()
	log.Println("TigerBeetle client closed")
}
