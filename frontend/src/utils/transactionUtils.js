import { HiCash, HiCreditCard, HiArrowRight } from 'react-icons/hi';

/**
 * Format currency amount to USD format
 * @param {number} amount - The amount to format
 * @returns {string} Formatted currency string
 */
export const formatCurrency = (amount) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
};

/**
 * Format date to a readable format
 * @param {string|Date} date - The date to format
 * @returns {string} Formatted date string
 */
export const formatDate = (date) => {
  const d = new Date(date);

  // Format: "Jan 15, 2025 at 3:45 PM"
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }) + ' at ' + d.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
};

/**
 * Format date to relative time (e.g., "2 hours ago")
 * @param {string|Date} date - The date to format
 * @returns {string} Relative time string
 */
export const formatRelativeTime = (date) => {
  const d = new Date(date);
  const now = new Date();
  const diffMs = now - d;
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;

  return formatDate(date);
};

/**
 * Format timestamp with both relative and absolute time
 * @param {string|Date} date - The date to format
 * @returns {string} Combined format: "3 hours ago • 2:30 PM, Jan 15"
 */
export const formatDetailedTimestamp = (date) => {
  const d = new Date(date);
  const now = new Date();
  const diffMs = now - d;
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  // Get relative time
  let relativeTime;
  if (diffSecs < 60) relativeTime = 'Just now';
  else if (diffMins < 60) relativeTime = `${diffMins} min${diffMins > 1 ? 's' : ''} ago`;
  else if (diffHours < 24) relativeTime = `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  else if (diffDays < 7) relativeTime = `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
  else relativeTime = null;

  // Get absolute time (always show time and date)
  const timeStr = d.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });

  const dateStr = d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: now.getFullYear() !== d.getFullYear() ? 'numeric' : undefined,
  });

  // Combine both if we have relative time
  if (relativeTime) {
    return `${relativeTime} • ${timeStr}, ${dateStr}`;
  }

  return `${timeStr}, ${dateStr}`;
};

/**
 * Get icon component for transaction type
 * @param {string} type - Transaction type (deposit, withdraw, transfer)
 * @returns {React.Component} Icon component
 */
export const getTransactionIcon = (type) => {
  const icons = {
    deposit: HiCash,
    withdraw: HiCreditCard,
    transfer: HiArrowRight,
  };

  return icons[type.toLowerCase()] || HiCash;
};

/**
 * Get color for transaction type
 * @param {string} type - Transaction type
 * @returns {string} Color hex code
 */
export const getTransactionColor = (type) => {
  const colors = {
    deposit: '#10B981',   // Green
    withdraw: '#F59E0B',  // Orange
    transfer: '#3B82F6',  // Blue
  };

  return colors[type.toLowerCase()] || '#10B981';
};

/**
 * Get badge variant for transaction type
 * @param {string} type - Transaction type
 * @returns {object} Badge style object
 */
export const getTransactionBadgeStyle = (type) => {
  const styles = {
    deposit: {
      backgroundColor: 'rgba(16, 185, 129, 0.15)',
      borderColor: 'rgba(16, 185, 129, 0.4)',
      color: '#10B981',
    },
    withdraw: {
      backgroundColor: 'rgba(245, 158, 11, 0.15)',
      borderColor: 'rgba(245, 158, 11, 0.4)',
      color: '#F59E0B',
    },
    transfer: {
      backgroundColor: 'rgba(59, 130, 246, 0.15)',
      borderColor: 'rgba(59, 130, 246, 0.4)',
      color: '#3B82F6',
    },
  };

  return styles[type.toLowerCase()] || styles.deposit;
};

/**
 * Format transaction amount with sign
 * @param {string} type - Transaction type
 * @param {number} amount - Transaction amount in cents
 * @returns {string} Formatted amount with sign
 */
export const formatTransactionAmount = (type, amount) => {
  const sign = type.toLowerCase() === 'deposit' ? '+' : '-';
  // Backend sends amount in cents, convert to dollars
  const amountInDollars = amount / 100;
  return `${sign}${formatCurrency(amountInDollars)}`;
};

/**
 * Get display label for transaction type
 * @param {string} type - Transaction type
 * @returns {string} Display label
 */
export const getTransactionLabel = (type) => {
  const labels = {
    deposit: 'Deposit',
    withdraw: 'Withdrawal',
    transfer: 'Transfer',
  };

  return labels[type.toLowerCase()] || type;
};
