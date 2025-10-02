import axios from "axios";
import router from '../router'

const baseURL = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/+$/, '')

const api = axios.create({
    baseURL,
    headers: { 'Content-Type': 'application/json'}
})

api.interceptors.request.use(config => {
    const token = localStorage.getItem('ottermq_token');
    if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
});

api.interceptors.response.use(
    (response) => response,
    (err) => {
        if (err?.response?.status === 401) {
            localStorage.removeItem('ottermq_token');
            try { router?.push?.({ path: '/login', query: { redirect: location.pathname }}) 
            } catch { window.location.assign('/login'); }
        }
        return Promise.reject(err);
    }
);

export default api;
