import { useState, useRef, useEffect } from 'react';
import { HiPaperAirplane, HiRefresh, HiTrash } from 'react-icons/hi';
import Message from './Message';

const ChatInterface = ({
  messages,
  onSendMessage,
  onAction,
  loading,
  onClear,
  showQuickActions = true
}) => {
  const [input, setInput] = useState('');
  const messagesEndRef = useRef(null);
  const textareaRef = useRef(null);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [input]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (input.trim() && !loading) {
      onSendMessage(input.trim());
      setInput('');
    }
  };

  const handleKeyDown = (e) => {
    // Enter to send, Shift+Enter for new line
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const handleQuickAction = (action) => {
    if (!loading) {
      onSendMessage(action);
    }
  };

  const handleClear = () => {
    if (onClear && window.confirm('Are you sure you want to clear the chat history?')) {
      onClear();
      setInput('');
    }
  };

  return (
    <div className="chat-interface">
      {/* Header */}
      <div className="chat-header">
        <div className="chat-header-title">
          <h3>AI Banking Assistant</h3>
          <span className="chat-status">
            <span className="status-dot"></span>
            {loading ? 'Thinking...' : 'Online'}
          </span>
        </div>
        {onClear && messages.length > 1 && (
          <button
            className="chat-clear-btn"
            onClick={handleClear}
            disabled={loading}
            title="Clear chat"
          >
            <HiTrash />
          </button>
        )}
      </div>

      {/* Messages Container */}
      <div className="chat-messages">
        {messages.length === 0 ? (
          <div className="chat-empty">
            <div className="chat-empty-icon">
              <HiRefresh size={64} />
            </div>
            <h3>Start a conversation</h3>
            <p>Ask me anything about your account, transactions, or financial operations!</p>
          </div>
        ) : (
          <>
            {messages.map((message) => (
              <Message key={message.id} message={message} onAction={onAction} />
            ))}

            {/* Typing Indicator */}
            {loading && (
              <div className="message message-ai">
                <div className="message-avatar">
                  <HiRefresh className="message-avatar-icon message-avatar-ai spinning" />
                </div>
                <div className="message-content-wrapper">
                  <div className="message-bubble">
                    <div className="typing-indicator">
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Quick Actions */}
      {showQuickActions && !loading && (
        <div className="chat-quick-actions">
          <button
            className="quick-action-btn"
            onClick={() => handleQuickAction('Show my balance')}
          >
            Check Balance
          </button>
          <button
            className="quick-action-btn"
            onClick={() => handleQuickAction('Show my last 5 transactions')}
          >
            Recent Transactions
          </button>
          <button
            className="quick-action-btn"
            onClick={() => handleQuickAction('Help me understand my account')}
          >
            Account Info
          </button>
        </div>
      )}

      {/* Input Area */}
      <form onSubmit={handleSubmit} className="chat-input-container">
        <div className="chat-input-wrapper">
          <textarea
            ref={textareaRef}
            className="chat-input"
            placeholder="Type your message... (Enter to send, Shift+Enter for new line)"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={loading}
            rows={1}
            maxLength={2000}
          />
          <button
            type="submit"
            className="chat-send-btn"
            disabled={!input.trim() || loading}
            aria-label="Send message"
          >
            <HiPaperAirplane />
          </button>
        </div>
        <div className="chat-input-hint">
          <span>Tip: Ask me to transfer money, check balance, or view transaction history</span>
        </div>
      </form>
    </div>
  );
};

export default ChatInterface;
