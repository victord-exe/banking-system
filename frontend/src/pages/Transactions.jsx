import { useState } from 'react';
import { HiCash, HiCreditCard, HiArrowRight } from 'react-icons/hi';
import { useAuth } from '../context/AuthContext';
import { transactionAPI } from '../services/api';
import TransactionCard from '../components/TransactionCard';
import AmountInput from '../components/AmountInput';
import Alert from '../components/Alert';
import ConfirmModal from '../components/ConfirmModal';

const Transactions = () => {
  const { balance, fetchBalance } = useAuth();
  const [activeTab, setActiveTab] = useState('deposit');
  const [loading, setLoading] = useState(false);
  const [alert, setAlert] = useState(null);

  // Deposit state
  const [depositAmount, setDepositAmount] = useState('');

  // Withdraw state
  const [withdrawAmount, setWithdrawAmount] = useState('');
  const [showWithdrawConfirm, setShowWithdrawConfirm] = useState(false);

  // Transfer state
  const [transferAccountId, setTransferAccountId] = useState('');
  const [transferAmount, setTransferAmount] = useState('');
  const [showTransferConfirm, setShowTransferConfirm] = useState(false);

  const showAlert = (type, message) => {
    setAlert({ type, message });
  };

  const closeAlert = () => {
    setAlert(null);
  };

  // DEPOSIT LOGIC
  const handleDeposit = async () => {
    const amount = parseFloat(depositAmount);

    if (!amount || amount <= 0) {
      showAlert('error', 'Please enter a valid amount');
      return;
    }

    if (amount > 10000) {
      showAlert('error', 'Maximum deposit amount is $10,000');
      return;
    }

    setLoading(true);
    try {
      await transactionAPI.deposit(amount);
      await fetchBalance();
      showAlert('success', `Successfully deposited $${amount.toFixed(2)}! New balance: $${(balance + amount).toFixed(2)}`);
      setDepositAmount('');
    } catch (error) {
      showAlert('error', error.response?.data?.error || 'Failed to process deposit');
    } finally {
      setLoading(false);
    }
  };

  // WITHDRAW LOGIC
  const handleWithdrawClick = () => {
    const amount = parseFloat(withdrawAmount);

    if (!amount || amount <= 0) {
      showAlert('error', 'Please enter a valid amount');
      return;
    }

    if (amount > balance) {
      showAlert('error', 'Insufficient funds');
      return;
    }

    if (amount > 5000) {
      showAlert('error', 'Maximum withdrawal amount is $5,000');
      return;
    }

    setShowWithdrawConfirm(true);
  };

  const handleWithdrawConfirm = async () => {
    const amount = parseFloat(withdrawAmount);
    setLoading(true);

    try {
      await transactionAPI.withdraw(amount);
      await fetchBalance();
      showAlert('success', `Successfully withdrew $${amount.toFixed(2)}! New balance: $${(balance - amount).toFixed(2)}`);
      setWithdrawAmount('');
      setShowWithdrawConfirm(false);
    } catch (error) {
      showAlert('error', error.response?.data?.error || 'Failed to process withdrawal');
      setShowWithdrawConfirm(false);
    } finally {
      setLoading(false);
    }
  };

  // TRANSFER LOGIC
  const handleTransferClick = () => {
    const amount = parseFloat(transferAmount);

    if (!transferAccountId || transferAccountId.trim() === '') {
      showAlert('error', 'Please enter a destination account ID');
      return;
    }

    if (!amount || amount <= 0) {
      showAlert('error', 'Please enter a valid amount');
      return;
    }

    if (amount > balance) {
      showAlert('error', 'Insufficient funds');
      return;
    }

    if (amount > 10000) {
      showAlert('error', 'Maximum transfer amount is $10,000');
      return;
    }

    setShowTransferConfirm(true);
  };

  const handleTransferConfirm = async () => {
    const amount = parseFloat(transferAmount);
    setLoading(true);

    try {
      await transactionAPI.transfer(transferAccountId, amount);
      await fetchBalance();
      showAlert('success', `Successfully transferred $${amount.toFixed(2)} to account ${transferAccountId}!`);
      setTransferAccountId('');
      setTransferAmount('');
      setShowTransferConfirm(false);
    } catch (error) {
      showAlert('error', error.response?.data?.error || 'Failed to process transfer');
      setShowTransferConfirm(false);
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'deposit', label: 'Deposit', icon: HiCash },
    { id: 'withdraw', label: 'Withdraw', icon: HiCreditCard },
    { id: 'transfer', label: 'Transfer', icon: HiArrowRight },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>Transactions</h1>
        <p>Manage your funds with ease</p>
      </div>

      {alert && (
        <Alert
          type={alert.type}
          message={alert.message}
          onClose={closeAlert}
        />
      )}

      <div className="transactions-page">
        {/* Tabs */}
        <div className="transaction-tabs">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                className={`transaction-tab ${activeTab === tab.id ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <Icon size={20} />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>

        {/* Forms */}
        <div className="transaction-forms">
          {/* DEPOSIT FORM */}
          {activeTab === 'deposit' && (
            <TransactionCard title="Deposit Funds" icon={HiCash} color="#10B981">
              <div className="transaction-form">
                {/* Quick Amount Buttons */}
                <div className="form-group">
                  <label className="form-label">Quick Amounts</label>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '0.5rem', marginTop: '0.5rem' }}>
                    {[10, 50, 100, 500, 1000].map((quickAmount) => (
                      <button
                        key={quickAmount}
                        type="button"
                        onClick={() => setDepositAmount(quickAmount.toString())}
                        style={{
                          padding: '0.5rem',
                          borderRadius: '6px',
                          border: depositAmount === quickAmount.toString() ? '2px solid #10B981' : '1px solid rgba(255, 255, 255, 0.2)',
                          background: depositAmount === quickAmount.toString() ? 'rgba(16, 185, 129, 0.2)' : 'rgba(255, 255, 255, 0.05)',
                          color: depositAmount === quickAmount.toString() ? '#10B981' : 'rgba(255, 255, 255, 0.8)',
                          cursor: 'pointer',
                          fontSize: '0.875rem',
                          fontWeight: '500',
                          transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                          if (depositAmount !== quickAmount.toString()) {
                            e.target.style.background = 'rgba(255, 255, 255, 0.1)'
                            e.target.style.borderColor = 'rgba(255, 255, 255, 0.3)'
                          }
                        }}
                        onMouseLeave={(e) => {
                          if (depositAmount !== quickAmount.toString()) {
                            e.target.style.background = 'rgba(255, 255, 255, 0.05)'
                            e.target.style.borderColor = 'rgba(255, 255, 255, 0.2)'
                          }
                        }}
                      >
                        ${quickAmount}
                      </button>
                    ))}
                  </div>
                </div>

                <AmountInput
                  label="Amount"
                  value={depositAmount}
                  onChange={setDepositAmount}
                  max={10000}
                  placeholder="0.00"
                />

                <div className="form-actions">
                  <button
                    className="btn btn-deposit"
                    onClick={handleDeposit}
                    disabled={loading}
                  >
                    <HiCash size={20} />
                    {loading ? 'Processing...' : 'Deposit'}
                  </button>
                </div>
              </div>
            </TransactionCard>
          )}

          {/* WITHDRAW FORM */}
          {activeTab === 'withdraw' && (
            <TransactionCard title="Withdraw Funds" icon={HiCreditCard} color="#F59E0B">
              <div className="transaction-form">
                <div className="form-group">
                  <label className="form-label">Current Balance</label>
                  <div className="balance-display">
                    ${balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </div>
                </div>

                {/* Quick Amount Buttons */}
                <div className="form-group">
                  <label className="form-label">Quick Amounts</label>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '0.5rem', marginTop: '0.5rem' }}>
                    {[10, 50, 100, 500, 1000].map((quickAmount) => {
                      const isDisabled = quickAmount > balance || quickAmount > 5000;
                      return (
                        <button
                          key={quickAmount}
                          type="button"
                          onClick={() => !isDisabled && setWithdrawAmount(quickAmount.toString())}
                          disabled={isDisabled}
                          style={{
                            padding: '0.5rem',
                            borderRadius: '6px',
                            border: withdrawAmount === quickAmount.toString() ? '2px solid #F59E0B' : '1px solid rgba(255, 255, 255, 0.2)',
                            background: withdrawAmount === quickAmount.toString()
                              ? 'rgba(245, 158, 11, 0.2)'
                              : isDisabled
                                ? 'rgba(255, 255, 255, 0.02)'
                                : 'rgba(255, 255, 255, 0.05)',
                            color: withdrawAmount === quickAmount.toString()
                              ? '#F59E0B'
                              : isDisabled
                                ? 'rgba(255, 255, 255, 0.3)'
                                : 'rgba(255, 255, 255, 0.8)',
                            cursor: isDisabled ? 'not-allowed' : 'pointer',
                            fontSize: '0.875rem',
                            fontWeight: '500',
                            transition: 'all 0.2s',
                            opacity: isDisabled ? 0.5 : 1,
                          }}
                          onMouseEnter={(e) => {
                            if (!isDisabled && withdrawAmount !== quickAmount.toString()) {
                              e.target.style.background = 'rgba(255, 255, 255, 0.1)'
                              e.target.style.borderColor = 'rgba(255, 255, 255, 0.3)'
                            }
                          }}
                          onMouseLeave={(e) => {
                            if (!isDisabled && withdrawAmount !== quickAmount.toString()) {
                              e.target.style.background = 'rgba(255, 255, 255, 0.05)'
                              e.target.style.borderColor = 'rgba(255, 255, 255, 0.2)'
                            }
                          }}
                        >
                          ${quickAmount}
                        </button>
                      );
                    })}
                  </div>
                </div>

                <AmountInput
                  label="Amount"
                  value={withdrawAmount}
                  onChange={setWithdrawAmount}
                  max={Math.min(5000, balance)}
                  placeholder="0.00"
                />

                <div className="form-actions">
                  <button
                    className="btn btn-withdraw"
                    onClick={handleWithdrawClick}
                    disabled={loading}
                  >
                    <HiCreditCard size={20} />
                    {loading ? 'Processing...' : 'Withdraw'}
                  </button>
                </div>
              </div>
            </TransactionCard>
          )}

          {/* TRANSFER FORM */}
          {activeTab === 'transfer' && (
            <TransactionCard title="Transfer Funds" icon={HiArrowRight} color="#3B82F6">
              <div className="transaction-form">
                <div className="form-group">
                  <label className="form-label">Current Balance</label>
                  <div className="balance-display">
                    ${balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">
                    Destination Account ID
                    <span style={{
                      fontSize: '0.75rem',
                      color: 'rgba(255, 255, 255, 0.5)',
                      marginLeft: '0.5rem',
                      fontWeight: 'normal'
                    }}>
                      (TigerBeetle Account ID)
                    </span>
                  </label>
                  <input
                    type="text"
                    className="form-input"
                    value={transferAccountId}
                    onChange={(e) => setTransferAccountId(e.target.value)}
                    placeholder="Enter TigerBeetle account ID"
                  />
                </div>

                {/* Quick Amount Buttons */}
                <div className="form-group">
                  <label className="form-label">Quick Amounts</label>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '0.5rem', marginTop: '0.5rem' }}>
                    {[10, 50, 100, 500, 1000].map((quickAmount) => {
                      const isDisabled = quickAmount > balance || quickAmount > 10000;
                      return (
                        <button
                          key={quickAmount}
                          type="button"
                          onClick={() => !isDisabled && setTransferAmount(quickAmount.toString())}
                          disabled={isDisabled}
                          style={{
                            padding: '0.5rem',
                            borderRadius: '6px',
                            border: transferAmount === quickAmount.toString() ? '2px solid #3B82F6' : '1px solid rgba(255, 255, 255, 0.2)',
                            background: transferAmount === quickAmount.toString()
                              ? 'rgba(59, 130, 246, 0.2)'
                              : isDisabled
                                ? 'rgba(255, 255, 255, 0.02)'
                                : 'rgba(255, 255, 255, 0.05)',
                            color: transferAmount === quickAmount.toString()
                              ? '#3B82F6'
                              : isDisabled
                                ? 'rgba(255, 255, 255, 0.3)'
                                : 'rgba(255, 255, 255, 0.8)',
                            cursor: isDisabled ? 'not-allowed' : 'pointer',
                            fontSize: '0.875rem',
                            fontWeight: '500',
                            transition: 'all 0.2s',
                            opacity: isDisabled ? 0.5 : 1,
                          }}
                          onMouseEnter={(e) => {
                            if (!isDisabled && transferAmount !== quickAmount.toString()) {
                              e.target.style.background = 'rgba(255, 255, 255, 0.1)'
                              e.target.style.borderColor = 'rgba(255, 255, 255, 0.3)'
                            }
                          }}
                          onMouseLeave={(e) => {
                            if (!isDisabled && transferAmount !== quickAmount.toString()) {
                              e.target.style.background = 'rgba(255, 255, 255, 0.05)'
                              e.target.style.borderColor = 'rgba(255, 255, 255, 0.2)'
                            }
                          }}
                        >
                          ${quickAmount}
                        </button>
                      );
                    })}
                  </div>
                </div>

                <AmountInput
                  label="Amount"
                  value={transferAmount}
                  onChange={setTransferAmount}
                  max={Math.min(10000, balance)}
                  placeholder="0.00"
                />

                <div className="form-actions">
                  <button
                    className="btn btn-transfer"
                    onClick={handleTransferClick}
                    disabled={loading}
                  >
                    <HiArrowRight size={20} />
                    {loading ? 'Processing...' : 'Transfer'}
                  </button>
                </div>
              </div>
            </TransactionCard>
          )}
        </div>
      </div>

      {/* Withdraw Confirmation Modal */}
      <ConfirmModal
        isOpen={showWithdrawConfirm}
        onClose={() => setShowWithdrawConfirm(false)}
        onConfirm={handleWithdrawConfirm}
        title="Confirm Withdrawal"
        message="Are you sure you want to withdraw this amount?"
        details={[
          { label: 'Amount', value: `$${parseFloat(withdrawAmount || 0).toFixed(2)}` },
          { label: 'New Balance', value: `$${(balance - parseFloat(withdrawAmount || 0)).toFixed(2)}` },
        ]}
        loading={loading}
        confirmText="Withdraw"
        variant="warning"
      />

      {/* Transfer Confirmation Modal */}
      <ConfirmModal
        isOpen={showTransferConfirm}
        onClose={() => setShowTransferConfirm(false)}
        onConfirm={handleTransferConfirm}
        title="Confirm Transfer"
        message="Are you sure you want to transfer this amount?"
        details={[
          { label: 'To Account', value: transferAccountId },
          { label: 'Amount', value: `$${parseFloat(transferAmount || 0).toFixed(2)}` },
          { label: 'New Balance', value: `$${(balance - parseFloat(transferAmount || 0)).toFixed(2)}` },
        ]}
        loading={loading}
        confirmText="Transfer"
        variant="danger"
      />
    </div>
  );
};

export default Transactions;
