import { HiCurrencyDollar } from 'react-icons/hi';

const AmountInput = ({
  value,
  onChange,
  label,
  error,
  min = 0,
  max,
  disabled = false,
  placeholder = "0.00"
}) => {
  const handleChange = (e) => {
    const inputValue = e.target.value;

    // Allow empty value
    if (inputValue === '') {
      onChange('');
      return;
    }

    // Remove non-numeric characters except decimal point and commas
    const numericValue = inputValue.replace(/[^0-9.]/g, '');

    // Ensure only one decimal point
    const parts = numericValue.split('.');
    if (parts.length > 2) return;

    // Limit to 2 decimal places
    if (parts[1] && parts[1].length > 2) return;

    onChange(numericValue);
  };

  const formatDisplayValue = () => {
    if (value === '' || value === undefined || value === null) return '';

    // Add thousand separators for display
    const parts = value.split('.');
    const integerPart = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    const decimalPart = parts[1] !== undefined ? '.' + parts[1] : '';

    return integerPart + decimalPart;
  };

  // Validate the current value
  const isValid = value && parseFloat(value) > 0 && (!max || parseFloat(value) <= max);

  return (
    <div className="form-group">
      {label && (
        <label className="form-label">
          {label}
          <span style={{
            fontSize: '0.875rem',
            color: 'rgba(255, 255, 255, 0.6)',
            marginLeft: '0.5rem',
            fontWeight: 'normal'
          }}>
            (USD - United States Dollar)
          </span>
        </label>
      )}
      <div className={`amount-input-wrapper ${error ? 'error' : ''} ${isValid ? 'valid' : ''}`}>
        <div className="amount-input-icon">
          <HiCurrencyDollar size={20} />
        </div>
        <input
          type="text"
          className="amount-input"
          value={formatDisplayValue()}
          onChange={handleChange}
          disabled={disabled}
          placeholder={placeholder}
          inputMode="decimal"
        />
      </div>
      {error && <span className="form-error">{error}</span>}
      {!error && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem' }}>
          <span className="form-hint" style={{ fontSize: '0.8rem', color: 'rgba(255, 255, 255, 0.5)' }}>
            Enter amount in dollars (e.g., 1000 = $1,000.00)
          </span>
          {max && (
            <span className="form-hint" style={{ fontSize: '0.8rem' }}>
              Max: ${max.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default AmountInput;
