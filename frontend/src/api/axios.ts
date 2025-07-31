import axios from 'axios';
import { getAuth } from 'firebase/auth';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
});

api.interceptors.request.use(async (cfg) => {
  const user = getAuth().currentUser;
  if (user) {
    const token = await user.getIdToken();
    cfg.headers.Authorization = `Bearer ${token}`;
  }
  return cfg;
});

// Add response interceptor to handle rate limiting
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 429) {
      const retryAfter = error.response.data?.retry_after || 60;
      const message = `Too many requests. Please try again in ${retryAfter} seconds.`;
      
      // You can emit a custom event here for a global notification system
      window.dispatchEvent(new CustomEvent('rate-limit-error', { 
        detail: { message, retryAfter } 
      }));
      
      // Enhance the error object with rate limit info
      error.isRateLimitError = true;
      error.retryAfter = retryAfter;
      error.rateLimitMessage = message;
    }
    return Promise.reject(error);
  }
);

export default api;
