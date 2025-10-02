import React, { createContext, useContext, useState, useEffect } from 'react';
const AuthContext = createContext(undefined);
const AUTH_STORAGE_KEY = 'shadowapi_auth';
export function AuthProvider({ children }) {
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    // Load auth from sessionStorage on mount
    useEffect(() => {
        const storedAuth = sessionStorage.getItem(AUTH_STORAGE_KEY);
        if (storedAuth) {
            try {
                const authData = JSON.parse(storedAuth);
                // Check if token is expired
                if (authData.expiresAt && authData.expiresAt > Date.now()) {
                    setUser(authData);
                }
                else {
                    console.log('Stored auth token has expired');
                    sessionStorage.removeItem(AUTH_STORAGE_KEY);
                }
            }
            catch (error) {
                console.error('Failed to parse stored auth data:', error);
                sessionStorage.removeItem(AUTH_STORAGE_KEY);
            }
        }
        setIsLoading(false);
    }, []);
    const login = (email, accessToken, idToken, refreshToken, expiresIn = 3600) => {
        const expiresAt = Date.now() + (expiresIn * 1000);
        const userData = { email, accessToken, idToken, refreshToken, expiresAt };
        setUser(userData);
        // Store in sessionStorage instead of localStorage for better security
        sessionStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(userData));
    };
    const logout = async () => {
        if (user?.accessToken) {
            try {
                const zitadelUrl = import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me';
                // Revoke the access token
                await fetch(`${zitadelUrl}/oauth/v2/revoke`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: new URLSearchParams({
                        token: user.accessToken,
                        token_type_hint: 'access_token',
                    }).toString(),
                });
            }
            catch (error) {
                console.error('Failed to revoke access token:', error);
            }
        }
        setUser(null);
        sessionStorage.removeItem(AUTH_STORAGE_KEY);
    };
    const checkAuth = async () => {
        if (!user?.accessToken) {
            return false;
        }
        try {
            // Check if token is expired
            if (user.expiresAt && user.expiresAt <= Date.now()) {
                console.log('Access token has expired');
                // TODO: Try to refresh using refresh token if available
                logout();
                return false;
            }
            // Token is valid and not expired
            return true;
        }
        catch (error) {
            console.error('Auth validation failed:', error);
            logout();
            return false;
        }
    };
    const getAccessToken = () => {
        if (!user?.accessToken) {
            return null;
        }
        // Check if token is expired
        if (user.expiresAt && user.expiresAt <= Date.now()) {
            return null;
        }
        return user.accessToken;
    };
    const value = {
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        checkAuth,
        getAccessToken
    };
    return (<AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>);
}
export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
//# sourceMappingURL=AuthContext.jsx.map