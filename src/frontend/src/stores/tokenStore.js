import { defineStore } from "pinia";

export const useTokenStore = defineStore("token", {
    state: () => ({
        username: null,
        token: null
    }),

    actions: {
        setUsername(newUsername) {
            this.username = newUsername;
        },
        setToken(newToken) {
            this.token = newToken;
        },
        clearToken() {
            this.username = null;
            this.token = null;
        }
    }
});
