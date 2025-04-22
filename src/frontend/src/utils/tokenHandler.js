import { useTokenStore } from "src/stores/tokenStore";

export function getUsername() {
    const tokenStore = useTokenStore();
    if (tokenStore.username) {
        return tokenStore.username;
    } else {
        const localUsername = localStorage.getItem("username");
        if (localUsername) {
            tokenStore.setUsername(localUsername);
            return localUsername;
        } else {
            return "";
        }
    }
}

export function setUserName(newUsername) {
    const tokenStore = useTokenStore();
    tokenStore.setUsername(newUsername);
    localStorage.setItem("username", newUsername);
}

export function getToken() {
    const tokenStore = useTokenStore();
    if (tokenStore.token) {
        return tokenStore.token;
    } else {
        const localToken = localStorage.getItem("token");
        if (localToken) {
            tokenStore.setToken(localToken);
            return localToken;
        } else {
            return "";
        }
    }
}

export function setToken(newToken) {
    const tokenStore = useTokenStore();
    tokenStore.setToken(newToken);
    localStorage.setItem("token", newToken);
}

export function removeToken() {
    const tokenStore = useTokenStore();
    tokenStore.clearToken();
    localStorage.removeItem("username");
    localStorage.removeItem("token");
}
