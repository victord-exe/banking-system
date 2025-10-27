import { useState, useEffect } from 'react';
import ChatInterface from '../components/ChatInterface';
import Alert from '../components/Alert';
import { chatAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const Chat = () => {
  const { fetchBalance } = useAuth();
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

      // DEBUG: Log the complete response
      console.log('🔍 FRONTEND: Full API response:', response);
      console.log('🔍 FRONTEND: response.data:', response.data);

      // Extract response data from the new backend structure
      const responseData = response.data.data;
      console.log('🔍 FRONTEND: Extracted responseData:', responseData);
      console.log('🔍 FRONTEND: responseData.reply:', responseData?.reply);
      console.log('🔍 FRONTEND: responseData.data:', responseData?.data);

      // Create AI response message
      const aiMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: responseData?.reply || 'I received your message.',
        timestamp: new Date(),
        // Add confirmation actions if needed
        actions: responseData?.requires_confirmation
          ? [
              {
                label: 'Confirm',
                type: 'confirm',
                toolName: responseData.confirmation_data?.tool_name,
                arguments: responseData.confirmation_data?.arguments,
              },
              {
                label: 'Cancel',
                type: 'cancel',
              },
            ]
          : null,
      };

      console.log('🔍 FRONTEND: Created aiMessage:', aiMessage);

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

  const handleAction = async (action) => {
    // Handle cancel action
    if (action.type === 'cancel') {
      const cancelMessage = {
        id: Date.now(),
        role: 'assistant',
        content: 'Operation cancelled.',
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, cancelMessage]);
      return;
    }

    // Handle confirm action
    if (action.type === 'confirm') {
      setLoading(true);
      setError(null);

      try {
        const response = await chatAPI.confirmOperation(
          action.toolName,
          action.arguments,
          true
        );

        const responseData = response.data.data;
        console.log('🟣 [Chat] Operation confirmed successfully:', responseData);
        const confirmMessage = {
          id: Date.now(),
          role: 'assistant',
          content: responseData?.reply || 'Operation completed.',
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, confirmMessage]);

        // Update balance after successful operation
        console.log('🟣 [Chat] Updating balance after operation...');
        await fetchBalance();
        console.log('🟣 [Chat] Balance updated after operation');
      } catch (err) {
        console.error('Confirmation error:', err);

        const errorMessage = {
          id: Date.now(),
          role: 'assistant',
          content: `I apologize, but I encountered an error: ${
            err.response?.data?.error ||
            err.response?.data?.message ||
            'Unable to complete the operation. Please try again.'
          }`,
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, errorMessage]);
        setError(
          err.response?.data?.error ||
          err.response?.data?.message ||
          'Failed to confirm operation. Please try again.'
        );
      } finally {
        setLoading(false);
      }
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
          onAction={handleAction}
          loading={loading}
          onClear={handleClear}
          showQuickActions={true}
        />
      </div>
    </div>
  );
};

export default Chat;
