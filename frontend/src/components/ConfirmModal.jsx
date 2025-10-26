import { HiX, HiExclamationCircle } from 'react-icons/hi';

const ConfirmModal = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  details = [],
  loading = false,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'warning' // warning | danger
}) => {
  if (!isOpen) return null;

  const variantColors = {
    warning: '#F59E0B',
    danger: '#EF4444',
  };

  const color = variantColors[variant];

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose} disabled={loading}>
          <HiX size={24} />
        </button>

        <div className="modal-header">
          <div className="modal-icon" style={{ color }}>
            <HiExclamationCircle size={48} />
          </div>
          <h2 className="modal-title">{title}</h2>
        </div>

        <div className="modal-body">
          <p className="modal-message">{message}</p>

          {details.length > 0 && (
            <div className="modal-details">
              {details.map((detail, index) => (
                <div key={index} className="modal-detail-item">
                  <span className="modal-detail-label">{detail.label}:</span>
                  <span className="modal-detail-value">{detail.value}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="modal-actions">
          <button
            className="btn btn-secondary"
            onClick={onClose}
            disabled={loading}
          >
            {cancelText}
          </button>
          <button
            className="btn btn-primary"
            onClick={onConfirm}
            disabled={loading}
            style={{ backgroundColor: color }}
          >
            {loading ? 'Processing...' : confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
