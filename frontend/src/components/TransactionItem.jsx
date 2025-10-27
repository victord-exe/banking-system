import {
  formatDetailedTimestamp,
  formatTransactionAmount,
  getTransactionIcon,
  getTransactionColor,
  getTransactionBadgeStyle,
  getTransactionLabel,
} from '../utils/transactionUtils';

const TransactionItem = ({ transaction }) => {
  const Icon = getTransactionIcon(transaction.type);
  const color = getTransactionColor(transaction.type);
  const badgeStyle = getTransactionBadgeStyle(transaction.type);
  const amountColor = transaction.type.toLowerCase() === 'deposit' ? '#10B981' : '#EF4444';

  // Determine recipient display based on transaction type
  const getRecipientDisplay = () => {
    if (transaction.type.toLowerCase() === 'transfer') {
      if (transaction.recipient_name || transaction.recipient_email) {
        const name = transaction.recipient_name || 'Unknown User';
        const email = transaction.recipient_email;
        return email ? `${name} (${email})` : name;
      }
      // Fallback to account ID if no user info
      return `Account ${transaction.credit_account_id}`;
    }
    return null;
  };

  const recipientDisplay = getRecipientDisplay();

  return (
    <div className="transaction-item">
      <div className="transaction-item-icon" style={{ color }}>
        <Icon size={24} />
      </div>

      <div className="transaction-item-content">
        <div className="transaction-item-header">
          <div className="transaction-item-type">
            <span className="transaction-label">{getTransactionLabel(transaction.type)}</span>
            <span
              className="transaction-badge"
              style={badgeStyle}
            >
              {transaction.type}
            </span>
          </div>
          <div className="transaction-item-amount" style={{ color: amountColor }}>
            {formatTransactionAmount(transaction.type, transaction.amount)}
          </div>
        </div>

        <div className="transaction-item-details">
          {/* Timestamp with both relative and absolute time */}
          <div className="transaction-detail">
            <span className="detail-label">Time:</span>
            <span className="detail-value">{formatDetailedTimestamp(transaction.created_at)}</span>
          </div>

          {/* Show recipient for transfers */}
          {recipientDisplay && (
            <div className="transaction-detail">
              <span className="detail-label">To:</span>
              <span className="detail-value">{recipientDisplay}</span>
            </div>
          )}

          {/* Show description if available */}
          {transaction.description && (
            <div className="transaction-detail">
              <span className="detail-label">Note:</span>
              <span className="detail-value">{transaction.description}</span>
            </div>
          )}

          {/* Status indicator */}
          {transaction.status && (
            <div className="transaction-detail">
              <span className="detail-label">Status:</span>
              <span className={`status-indicator status-${transaction.status.toLowerCase()}`}>
                {transaction.status}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TransactionItem;
