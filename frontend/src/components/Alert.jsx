import { HiCheckCircle, HiXCircle, HiExclamationCircle, HiInformationCircle, HiX } from 'react-icons/hi';
import { useEffect } from 'react';

const Alert = ({ type = 'info', message, onClose, autoDismiss = true }) => {
  useEffect(() => {
    if (autoDismiss && onClose) {
      const timer = setTimeout(() => {
        onClose();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [autoDismiss, onClose]);

  const configs = {
    success: {
      icon: HiCheckCircle,
      bgColor: 'rgba(16, 185, 129, 0.1)',
      borderColor: '#10B981',
      textColor: '#10B981',
      iconColor: '#10B981',
    },
    error: {
      icon: HiXCircle,
      bgColor: 'rgba(239, 68, 68, 0.1)',
      borderColor: '#EF4444',
      textColor: '#EF4444',
      iconColor: '#EF4444',
    },
    warning: {
      icon: HiExclamationCircle,
      bgColor: 'rgba(245, 158, 11, 0.1)',
      borderColor: '#F59E0B',
      textColor: '#F59E0B',
      iconColor: '#F59E0B',
    },
    info: {
      icon: HiInformationCircle,
      bgColor: 'rgba(59, 130, 246, 0.1)',
      borderColor: '#3B82F6',
      textColor: '#3B82F6',
      iconColor: '#3B82F6',
    },
  };

  const config = configs[type];
  const Icon = config.icon;

  return (
    <div
      className="alert"
      style={{
        backgroundColor: config.bgColor,
        borderLeft: `4px solid ${config.borderColor}`,
        color: config.textColor,
      }}
    >
      <div className="alert-icon" style={{ color: config.iconColor }}>
        <Icon size={24} />
      </div>
      <div className="alert-message">{message}</div>
      {onClose && (
        <button
          className="alert-close"
          onClick={onClose}
          aria-label="Close alert"
        >
          <HiX size={20} />
        </button>
      )}
    </div>
  );
};

export default Alert;
