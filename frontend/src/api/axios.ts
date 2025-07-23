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

export default api;
