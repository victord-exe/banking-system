package chat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// intentPattern defines a pattern for detecting user intents
type intentPattern struct {
	intent   Intent
	patterns []*regexp.Regexp
}

var (
	// Regex patterns for intent detection
	intentPatterns = []intentPattern{
		// Balance queries
		{
			intent: IntentBalance,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(what'?s?\s+my|show\s+my|check\s+my|get\s+my)?\s*balance`),
				regexp.MustCompile(`(?i)how\s+much\s+(money|funds|cash)(\s+do\s+i\s+have)?`),
				regexp.MustCompile(`(?i)account\s+balance`),
			},
		},
		// Deposit operations
		{
			intent: IntentDeposit,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)deposit\s+`),
				regexp.MustCompile(`(?i)add\s+.*\s+(to\s+my\s+account|to\s+account)`),
				regexp.MustCompile(`(?i)put\s+.*\s+(in|into)\s+(my\s+)?account`),
			},
		},
		// Withdrawal operations
		{
			intent: IntentWithdraw,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)withdraw\s+`),
				regexp.MustCompile(`(?i)(take\s+out|remove)\s+`),
				regexp.MustCompile(`(?i)cash\s+out\s+`),
			},
		},
		// Transfer operations
		{
			intent: IntentTransfer,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)transfer\s+.*\s+to\s+`),
				regexp.MustCompile(`(?i)send\s+.*\s+to\s+(account\s+)?`),
				regexp.MustCompile(`(?i)pay\s+.*\s+to\s+`),
			},
		},
		// Transaction history
		{
			intent: IntentHistory,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(show|get|view)\s+(my\s+)?(transaction\s+)?histor(y|ies)`),
				regexp.MustCompile(`(?i)(last|recent)\s+\d*\s*transactions?`),
				regexp.MustCompile(`(?i)transaction\s+(list|log)`),
				regexp.MustCompile(`(?i)my\s+transactions?`),
			},
		},
	}

	// Regex for extracting amounts (supports $100, 100, $100.50, 100.50)
	amountPattern = regexp.MustCompile(`(?i)\$?\s*(\d+(?:\.\d{1,2})?)`)

	// Regex for extracting account IDs in transfers
	accountIDPattern = regexp.MustCompile(`(?i)(?:to\s+)?(?:account\s+)?(\d+)`)

	// Regex for extracting limits in history queries
	limitPattern = regexp.MustCompile(`(?i)(last|recent)\s+(\d+)`)
)

// ParseIntent analyzes a user message and extracts intent and parameters
func ParseIntent(message string) ParsedIntent {
	message = strings.TrimSpace(message)

	// Detect intent
	intent := detectIntent(message)

	result := ParsedIntent{
		Intent: intent,
	}

	// Extract parameters based on intent
	switch intent {
	case IntentDeposit, IntentWithdraw:
		result.Amount = extractAmount(message)

	case IntentTransfer:
		result.Amount = extractAmount(message)
		result.ToAccountID = extractAccountID(message)

	case IntentHistory:
		result.Limit = extractLimit(message)
		if result.Limit == 0 {
			result.Limit = 10 // Default to 10 transactions
		}
	}

	return result
}

// detectIntent identifies the user's intent from the message
func detectIntent(message string) Intent {
	for _, ip := range intentPatterns {
		for _, pattern := range ip.patterns {
			if pattern.MatchString(message) {
				return ip.intent
			}
		}
	}

	return IntentUnknown
}

// extractAmount extracts a monetary amount from the message
// Returns amount in cents (e.g., "100.50" becomes 10050)
func extractAmount(message string) int64 {
	matches := amountPattern.FindStringSubmatch(message)
	if len(matches) < 2 {
		return 0
	}

	// Parse the amount as float
	amountStr := matches[1]
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0
	}

	// Convert to cents (multiply by 100)
	return int64(amount * 100)
}

// extractAccountID extracts a destination account ID from the message
func extractAccountID(message string) uint64 {
	matches := accountIDPattern.FindStringSubmatch(message)
	if len(matches) < 2 {
		return 0
	}

	accountID, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0
	}

	return accountID
}

// extractLimit extracts the number of transactions to show from history queries
func extractLimit(message string) int {
	matches := limitPattern.FindStringSubmatch(message)
	if len(matches) < 3 {
		return 0
	}

	limit, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0
	}

	// Cap at reasonable maximum
	if limit > 100 {
		limit = 100
	}

	return limit
}

// ValidateIntent checks if the parsed intent has all required parameters
func ValidateIntent(parsed ParsedIntent) error {
	switch parsed.Intent {
	case IntentDeposit, IntentWithdraw:
		if parsed.Amount <= 0 {
			return fmt.Errorf("please specify a valid amount (e.g., '$100' or '50.25')")
		}

	case IntentTransfer:
		if parsed.Amount <= 0 {
			return fmt.Errorf("please specify a valid transfer amount")
		}
		if parsed.ToAccountID == 0 {
			return fmt.Errorf("please specify the destination account ID (e.g., 'to account 12345')")
		}

	case IntentUnknown:
		return fmt.Errorf("I didn't understand that request")
	}

	return nil
}
