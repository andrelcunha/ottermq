import axios from "axios";

const baseURL = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/+$/, '')

const api = axios.create({
    baseURL,
    headers: { 'Content-Type': 'application/json'}
})

api.interceptors.request.use(config => {
    const token = localStorage.getItem('ottermq_token');
    if (token) config.headers['Authorization'] = `Bearer ${token}`;
    return config;
});

api.interceptors.response.use(
    (response) => response,
    (err) => {
        // optional: kick to login on 401
        if (err?.response?.status === 401) {
            // localStorage.removeItem('ottermq_token');
        }
        return Promise.reject(err);
    }
);

export default api;
