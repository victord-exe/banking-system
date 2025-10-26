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

    // Remove non-numeric characters except decimal point
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
    return value;
  };

  return (
    <div className="form-group">
      {label && <label className="form-label">{label}</label>}
      <div className={`amount-input-wrapper ${error ? 'error' : ''}`}>
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
      {max && !error && (
        <span className="form-hint">Maximum: ${max.toLocaleString()}</span>
      )}
    </div>
  );
};

export default AmountInput;
