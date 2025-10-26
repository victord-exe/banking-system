import {
  formatRelativeTime,
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
          <div className="transaction-detail">
            <span className="detail-label">Time:</span>
            <span className="detail-value">{formatRelativeTime(transaction.created_at)}</span>
          </div>

          {transaction.from_account_id && (
            <div className="transaction-detail">
              <span className="detail-label">From:</span>
              <span className="detail-value mono">{transaction.from_account_id}</span>
            </div>
          )}

          {transaction.to_account_id && (
            <div className="transaction-detail">
              <span className="detail-label">To:</span>
              <span className="detail-value mono">{transaction.to_account_id}</span>
            </div>
          )}

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
