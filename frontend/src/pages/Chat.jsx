import { useState, useEffect } from 'react';
import ChatInterface from '../components/ChatInterface';
import Alert from '../components/Alert';
import { chatAPI } from '../services/api';

const Chat = () => {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Initialize with welcome message
  useEffect(() => {
    const welcomeMessage = {
      id: Date.now(),
      role: 'assistant',
      content: 'Hello! I\'m your AI Banking Assistant. I can help you with:\n\n• Checking your account balance\n• Viewing transaction history\n• Making deposits, withdrawals, and transfers\n• Understanding your account details\n\nWhat would you like to do today?',
      timestamp: new Date(),
    };
    setMessages([welcomeMessage]);
  }, []);

  const handleSendMessage = async (content) => {
    // Create user message
    const userMessage = {
      id: Date.now(),
      role: 'user',
      content,
      timestamp: new Date(),
    };

    // Add user message immediately
    setMessages((prev) => [...prev, userMessage]);
    setError(null);
    setLoading(true);

    try {
      // Call API
      const response = await chatAPI.sendMessage(content);

      // Create AI response message
      const aiMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: response.data.message || response.data.response || 'I received your message.',
        timestamp: new Date(),
        // If the response includes actions (like confirm transfer), add them
        actions: response.data.actions || null,
      };

      setMessages((prev) => [...prev, aiMessage]);
    } catch (err) {
      console.error('Chat error:', err);

      // Create error message
      const errorMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: `I apologize, but I encountered an error: ${
          err.response?.data?.error ||
          err.response?.data?.message ||
          'Unable to process your request. Please try again.'
        }`,
        timestamp: new Date(),
      };

      setMessages((prev) => [...prev, errorMessage]);
      setError(
        err.response?.data?.error ||
        err.response?.data?.message ||
        'Failed to send message. Please try again.'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    const welcomeMessage = {
      id: Date.now(),
      role: 'assistant',
      content: 'Chat cleared. How can I help you today?',
      timestamp: new Date(),
    };
    setMessages([welcomeMessage]);
    setError(null);
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>AI Banking Assistant</h1>
        <p>Interact with your bank using natural language</p>
      </div>

      {error && (
        <Alert
          type="error"
          message={error}
          onClose={() => setError(null)}
        />
      )}

      <div className="chat-page-content">
        <ChatInterface
          messages={messages}
          onSendMessage={handleSendMessage}
          loading={loading}
          onClear={handleClear}
          showQuickActions={true}
        />
      </div>
    </div>
  );
};

export default Chat;
