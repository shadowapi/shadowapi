import { useState } from 'react';
import client from '../api/client';
import { config } from '../config/env';
export function useZitadelAuth() {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [fieldErrors, setFieldErrors] = useState({});
    const [authConfig, setAuthConfig] = useState(null);
    const parseZitadelError = async (response) => {
        try {
            const errorData = await response.json();
            let message = errorData.message || `Request failed with status ${response.status}`;
            const fieldErrors = {};
            // Extract field-specific errors from violations
            if (errorData.details) {
                for (const detail of errorData.details) {
                    if (detail.violations) {
                        for (const violation of detail.violations) {
                            if (violation.field && violation.description) {
                                // Map common Zitadel field names to our form fields
                                const fieldName = violation.field.toLowerCase().includes('loginname') ||
                                    violation.field.toLowerCase().includes('email') ? 'email' :
                                    violation.field.toLowerCase().includes('password') ? 'password' : violation.field;
                                fieldErrors[fieldName] = violation.description;
                            }
                        }
                    }
                }
            }
            // Common Zitadel error codes to user-friendly messages
            if (errorData.code) {
                switch (errorData.code) {
                    case 'PRECONDITION_FAILED':
                        message = 'Invalid email or password';
                        fieldErrors.email = 'Invalid email or password';
                        break;
                    case 'NOT_FOUND':
                        message = 'User not found';
                        fieldErrors.email = 'User not found';
                        break;
                    case 'INVALID_ARGUMENT':
                        message = 'Invalid credentials';
                        fieldErrors.email = 'Invalid credentials';
                        break;
                    default:
                        // Keep the original message or use a fallback
                        break;
                }
            }
            return { message, fieldErrors };
        }
        catch (parseError) {
            // Fallback if JSON parsing fails
            const text = await response.text();
            return {
                message: text || `Request failed with status ${response.status}`,
                fieldErrors: {}
            };
        }
    };
    const getSessionToken = async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await client.POST('/user/session', {});
            if (response.error) {
                throw new Error(response.error.detail || 'Failed to get session token');
            }
            if (!response.data) {
                throw new Error('No data received from backend');
            }
            const { session_token, zitadel_url, expires_in } = response.data;
            if (!session_token) {
                throw new Error('No session token received from backend');
            }
            if (!zitadel_url) {
                throw new Error('No Zitadel URL received from backend');
            }
            const config = {
                zitadelUrl: zitadel_url
            };
            setAuthConfig(config);
            console.log('Returning token data:', {
                sessionToken: session_token?.substring(0, 20) + '...',
                zitadelUrl: zitadel_url,
                expiresIn: expires_in
            });
            return {
                sessionToken: session_token,
                zitadelUrl: zitadel_url,
                expiresIn: expires_in
            };
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to get session token';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const createZitadelSession = async (loginName, zitadelUrl, bearerToken) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Creating Zitadel session:', {
                loginName,
                zitadelUrl,
                bearerToken: bearerToken ? `Bearer ${bearerToken.substring(0, 20)}...` : 'NO TOKEN',
                fullBearerToken: bearerToken
            });
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${bearerToken}`,
            };
            const response = await fetch(`${zitadelUrl}/v2/sessions`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    checks: {
                        user: {
                            loginName: loginName
                        }
                    }
                })
            });
            if (!response.ok) {
                const { message, fieldErrors } = await parseZitadelError(response);
                setFieldErrors(fieldErrors);
                throw new Error(message);
            }
            const sessionData = await response.json();
            return {
                sessionId: sessionData.sessionId,
                sessionToken: sessionData.sessionToken,
                changeDate: sessionData.changeDate
            };
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to create Zitadel session';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const addPasswordToSession = async (sessionId, password, zitadelUrl, bearerToken) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Adding password to session:', {
                sessionId,
                zitadelUrl,
                bearerToken: bearerToken ? `Bearer ${bearerToken.substring(0, 20)}...` : 'NO TOKEN',
                fullBearerToken: bearerToken
            });
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${bearerToken}`,
            };
            const response = await fetch(`${zitadelUrl}/v2/sessions/${sessionId}`, {
                method: 'PATCH',
                headers,
                body: JSON.stringify({
                    checks: {
                        password: {
                            password: password
                        }
                    }
                })
            });
            if (!response.ok) {
                const { message, fieldErrors } = await parseZitadelError(response);
                setFieldErrors(fieldErrors);
                throw new Error(message);
            }
            const sessionData = await response.json();
            return {
                sessionId: sessionData.sessionId,
                sessionToken: sessionData.sessionToken,
                changeDate: sessionData.changeDate
            };
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Password verification failed';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const createAuthRequest = async (zitadelUrl, bearerToken) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Creating OIDC auth request...');
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${bearerToken}`,
            };
            const response = await fetch(`${zitadelUrl}/v2/oidc/auth_requests`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    client_id: config.zitadel.clientId,
                    redirect_uri: config.zitadel.redirectUri,
                    scope: ['openid', 'profile', 'email'],
                    response_type: 'code',
                })
            });
            if (!response.ok) {
                const { message, fieldErrors } = await parseZitadelError(response);
                setFieldErrors(fieldErrors);
                throw new Error(message);
            }
            const authRequestData = await response.json();
            console.log('Auth request created:', authRequestData.authRequestId);
            return authRequestData.authRequestId;
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to create auth request';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const finalizeAuthRequest = async (authRequestId, sessionId, sessionToken, zitadelUrl, bearerToken) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Finalizing auth request with session...');
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${bearerToken}`,
            };
            const response = await fetch(`${zitadelUrl}/v2/oidc/auth_requests/${authRequestId}`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    session: {
                        sessionId,
                        sessionToken,
                    }
                })
            });
            if (!response.ok) {
                const { message, fieldErrors } = await parseZitadelError(response);
                setFieldErrors(fieldErrors);
                throw new Error(message);
            }
            const finalizeData = await response.json();
            console.log('Auth request finalized, got callback URL');
            // Extract authorization code from callback URL
            const callbackUrl = new URL(finalizeData.callbackUrl);
            const code = callbackUrl.searchParams.get('code');
            if (!code) {
                throw new Error('No authorization code in callback URL');
            }
            return code;
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to finalize auth request';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const exchangeCodeForTokens = async (code, zitadelUrl) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Exchanging authorization code for tokens...');
            const params = new URLSearchParams({
                grant_type: 'authorization_code',
                code,
                client_id: config.zitadel.clientId,
                redirect_uri: config.zitadel.redirectUri,
            });
            const response = await fetch(`${zitadelUrl}/oauth/v2/token`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: params.toString(),
            });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Token exchange failed: ${errorText}`);
            }
            const tokens = await response.json();
            console.log('Successfully obtained OIDC tokens');
            return tokens;
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to exchange code for tokens';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const getOIDCTokensFromSession = async (sessionId, sessionToken, zitadelUrl, bearerToken) => {
        setLoading(true);
        setError(null);
        try {
            console.log('Step 4: Initiating OIDC authorization flow...');
            // Build authorize URL
            const authorizeParams = new URLSearchParams({
                client_id: config.zitadel.clientId,
                redirect_uri: config.zitadel.redirectUri,
                response_type: 'code',
                scope: 'openid profile email',
                prompt: 'none', // Don't show login UI since we have a session
            });
            // Step 1: Call authorize endpoint to create auth request
            // This will redirect, but we'll capture the authRequest ID from the redirect
            const authorizeUrl = `${zitadelUrl}/oauth/v2/authorize?${authorizeParams.toString()}`;
            console.log('Calling authorize endpoint:', authorizeUrl);
            const authorizeResponse = await fetch(authorizeUrl, {
                method: 'GET',
                redirect: 'manual', // Don't follow redirects automatically
                credentials: 'include', // Include cookies
            });
            // Extract authRequestId from Location header
            const location = authorizeResponse.headers.get('location');
            if (!location) {
                throw new Error('No redirect location from authorize endpoint');
            }
            console.log('Got redirect location:', location);
            // Parse authRequest ID from redirect URL
            const redirectUrl = new URL(location, zitadelUrl);
            const authRequestId = redirectUrl.searchParams.get('authRequest');
            if (!authRequestId) {
                // If no authRequest param, check if we got a code directly (session was recognized)
                const code = redirectUrl.searchParams.get('code');
                if (code) {
                    console.log('Got authorization code directly (session recognized)');
                    return await exchangeCodeForTokens(code, zitadelUrl);
                }
                throw new Error('No authRequest ID or code in authorize redirect');
            }
            console.log('Got authRequest ID:', authRequestId);
            // Step 2: Finalize auth request with our session
            const finalizeResponse = await fetch(`${zitadelUrl}/v2/oidc/auth_requests/${authRequestId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${bearerToken}`,
                },
                body: JSON.stringify({
                    session: {
                        sessionId,
                        sessionToken,
                    }
                })
            });
            if (!finalizeResponse.ok) {
                const { message, fieldErrors } = await parseZitadelError(finalizeResponse);
                setFieldErrors(fieldErrors);
                throw new Error(message);
            }
            const finalizeData = await finalizeResponse.json();
            console.log('Auth request finalized, got callback URL');
            // Step 3: Extract authorization code from callback URL
            const callbackUrl = new URL(finalizeData.callbackUrl);
            const code = callbackUrl.searchParams.get('code');
            if (!code) {
                throw new Error('No authorization code in callback URL');
            }
            // Step 4: Exchange code for tokens
            return await exchangeCodeForTokens(code, zitadelUrl);
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to get OIDC tokens';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    const authenticateWithZitadel = async (email, password) => {
        setLoading(true);
        setError(null);
        setFieldErrors({});
        try {
            // Step 1: Get session token from backend
            const tokenData = await getSessionToken();
            // Step 2: Create Zitadel session with username
            const zitadelLoginName = email;
            const session = await createZitadelSession(zitadelLoginName, tokenData.zitadelUrl, tokenData.sessionToken);
            // Step 3: Add password to session
            console.log('Step 3: Adding password to session...');
            const authenticatedSession = await addPasswordToSession(session.sessionId, password, tokenData.zitadelUrl, tokenData.sessionToken);
            // Step 4-7: Get OIDC tokens from the authenticated session
            const tokens = await getOIDCTokensFromSession(authenticatedSession.sessionId, authenticatedSession.sessionToken, tokenData.zitadelUrl, tokenData.sessionToken);
            return tokens;
        }
        catch (err) {
            const message = err instanceof Error ? err.message : 'Authentication failed';
            setError(message);
            throw new Error(message);
        }
        finally {
            setLoading(false);
        }
    };
    return {
        loading,
        error,
        fieldErrors,
        authConfig,
        getSessionToken,
        createZitadelSession,
        addPasswordToSession,
        authenticateWithZitadel
    };
}
//# sourceMappingURL=useZitadelAuth.js.map