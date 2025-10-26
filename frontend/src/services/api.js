import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 errors (logout)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/';
    }
    return Promise.reject(error);
  }
);

// Auth endpoints
export const authAPI = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  logout: () => api.post('/auth/logout'),
};

// Account endpoints
export const accountAPI = {
  getBalance: () => api.get('/accounts/balance'),
  getMe: () => api.get('/accounts/me'),
};

// Transaction endpoints
export const transactionAPI = {
  deposit: (amount) => api.post('/transactions/deposit', { amount }),
  withdraw: (amount) => api.post('/transactions/withdraw', { amount }),
  transfer: (toAccountId, amount) => api.post('/transactions/transfer', { to_account_id: toAccountId, amount }),
  getHistory: (page = 1, limit = 10) => api.get(`/transactions/history?page=${page}&limit=${limit}`),
};

// Chat endpoints
export const chatAPI = {
  sendMessage: (message) => api.post('/chat', { message }),
};

export default api;