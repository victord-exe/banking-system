import { useState, useEffect } from 'react';
import { HiRefresh, HiDocumentText } from 'react-icons/hi';
import { transactionAPI } from '../services/api';
import TransactionList from '../components/TransactionList';
import Pagination from '../components/Pagination';
import Alert from '../components/Alert';

const History = () => {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10);
  const [totalTransactions, setTotalTransactions] = useState(0);

  const totalPages = Math.ceil(totalTransactions / itemsPerPage);

  const fetchTransactions = async (page = currentPage, limit = itemsPerPage) => {
    setLoading(true);
    setError(null);

    try {
      const response = await transactionAPI.getHistory(page, limit);

      // Backend wraps response in a "data" object
      const responseData = response.data.data;

      if (responseData && responseData.transactions) {
        setTransactions(responseData.transactions);
        setTotalTransactions(responseData.pagination?.total || responseData.transactions.length);
      } else {
        console.warn('Unexpected API response format:', response.data);
        setTransactions([]);
        setTotalTransactions(0);
      }
    } catch (err) {
      console.error('Error fetching transactions:', err);
      setError(err.response?.data?.error || 'Failed to load transaction history');
      setTransactions([]);
      setTotalTransactions(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTransactions(currentPage, itemsPerPage);
  }, [currentPage, itemsPerPage]);

  const handlePageChange = (newPage) => {
    setCurrentPage(newPage);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1); // Reset to first page when changing items per page
  };

  const handleRefresh = () => {
    fetchTransactions(currentPage, itemsPerPage);
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-content">
          <div className="page-title-section">
            <HiDocumentText size={36} className="page-icon" />
            <div>
              <h1>Transaction History</h1>
              <p>View and track all your past transactions</p>
            </div>
          </div>
          <button
            className="btn-refresh"
            onClick={handleRefresh}
            disabled={loading}
            aria-label="Refresh transactions"
          >
            <HiRefresh size={20} className={loading ? 'spinning' : ''} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {error && (
        <Alert
          type="error"
          message={error}
          onClose={() => setError(null)}
        />
      )}

      <div className="history-stats">
        <div className="stat-card">
          <span className="stat-label">Total Transactions</span>
          <span className="stat-value">{totalTransactions}</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Current Page</span>
          <span className="stat-value">{currentPage} of {totalPages || 1}</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Showing</span>
          <span className="stat-value">{itemsPerPage} per page</span>
        </div>
      </div>

      <div className="history-content">
        <TransactionList
          transactions={transactions}
          loading={loading}
        />

        {!loading && transactions.length > 0 && (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            itemsPerPage={itemsPerPage}
            totalItems={totalTransactions}
            onPageChange={handlePageChange}
            onItemsPerPageChange={handleItemsPerPageChange}
          />
        )}
      </div>
    </div>
  );
};

export default History;
