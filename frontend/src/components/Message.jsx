import { HiUser, HiSparkles } from 'react-icons/hi';
import { formatDistanceToNow } from 'date-fns';

const Message = ({ message, onAction }) => {
  const { role, content, timestamp, actions } = message;

  const isUser = role === 'user';
  const isSystem = role === 'system';

  const formatTimestamp = (ts) => {
    if (!ts) return '';
    try {
      const date = typeof ts === 'string' ? new Date(ts) : ts;
      return formatDistanceToNow(date, { addSuffix: true });
    } catch (error) {
      return '';
    }
  };

  if (isSystem) {
    return (
      <div className="message-system">
        <div className="message-system-content">
          <div className="message-system-icon">
            <HiSparkles />
          </div>
          <p>{content}</p>
        </div>
        {timestamp && (
          <div className="message-timestamp">{formatTimestamp(timestamp)}</div>
        )}
      </div>
    );
  }

  return (
    <div className={`message ${isUser ? 'message-user' : 'message-ai'}`}>
      <div className="message-avatar">
        {isUser ? (
          <HiUser className="message-avatar-icon" />
        ) : (
          <HiSparkles className="message-avatar-icon message-avatar-ai" />
        )}
      </div>

      <div className="message-content-wrapper">
        <div className="message-bubble">
          <div className="message-content">
            {content}
          </div>

          {actions && actions.length > 0 && (
            <div className="message-actions">
              {actions.map((action, index) => (
                <button
                  key={index}
                  className="message-action-btn"
                  onClick={() => onAction && onAction(action)}
                  disabled={action.disabled}
                >
                  {action.label}
                </button>
              ))}
            </div>
          )}
        </div>

        {timestamp && (
          <div className="message-timestamp">{formatTimestamp(timestamp)}</div>
        )}
      </div>
    </div>
  );
};

export default Message;
