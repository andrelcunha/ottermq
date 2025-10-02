import { defineStore } from 'pinia';
import api from 'src/services/api';

export const useAuthStore = defineStore('auth', {
    state: () => ({
        token: localStorage.getItem('ottermq_token') || '',
        username: '',
        loading: false,
        error: null,
    }),
    getters: {
        isAuthed: (state) => !!state.token,
    },
    actions: {
        async login(username, password) {
            this.loading = true;
            this.error = null;
            try {
                const { data } = await api.post('/login', { username, password });
                const token = data?.token;
                if (!token) throw new Error('Token missing');
                this.token = token;
                this.username = username;
                localStorage.setItem('ottermq_token', token);
            } catch (e) {
                this.error = e?.response?.data?.error || e.message;
                throw e;
            } finally {
                this.loading = false;
            }
        },
        logout() {
            this.token = '';
            this.username = '';
            localStorage.removeItem('ottermq_token');
        },
    },
});