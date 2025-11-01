// Utility functions for authentication and authorization

export const getCookie = (name) => {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
};

export const setCookie = (name, value, days = 1) => {
    const expires = new Date();
    expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
    document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/`;
};

export const deleteCookie = (name) => {
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;`;
};

export const decodeJWT = (token) => {
    try {
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(
            atob(base64)
                .split('')
                .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
                .join('')
        );
        return JSON.parse(jsonPayload);
    } catch (error) {
        console.error('Error decoding JWT:', error);
        return null;
    }
};

export const getAuthToken = () => {
    return getCookie('jwt');
};

export const getUserFromToken = () => {
    const token = getAuthToken();
    if (!token) return null;
    
    const decoded = decodeJWT(token);
    return decoded;
};

export const isAuthenticated = () => {
    const token = getAuthToken();
    if (!token) return false;
    
    const decoded = decodeJWT(token);
    if (!decoded || !decoded.exp) return false;
    
    // Check if token is expired
    const currentTime = Date.now() / 1000;
    return decoded.exp > currentTime;
};

export const isAdmin = () => {
    const user = getUserFromToken();
    return user && user.role === 'admin';
};

export const logout = () => {
    deleteCookie('jwt');
};

