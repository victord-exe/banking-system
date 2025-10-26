import { HiReceiptTax } from 'react-icons/hi';
import TransactionItem from './TransactionItem';

const TransactionList = ({ transactions, loading }) => {
  if (loading) {
    return (
      <div className="transaction-list-loading">
        <div className="loading-spinner"></div>
        <p>Loading transactions...</p>
      </div>
    );
  }

  if (!transactions || transactions.length === 0) {
    return (
      <div className="transaction-list-empty">
        <div className="empty-icon">
          <HiReceiptTax size={64} />
        </div>
        <h3>No Transactions Yet</h3>
        <p>Your transaction history will appear here once you make your first transaction.</p>
        <p className="empty-hint">Try making a deposit or transfer to get started!</p>
      </div>
    );
  }

  return (
    <div className="transaction-list">
      {transactions.map((transaction) => (
        <TransactionItem
          key={transaction.id}
          transaction={transaction}
        />
      ))}
    </div>
  );
};

export default TransactionList;
