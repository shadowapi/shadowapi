# ZITADEL Username/Password Authentication

This document explains how to implement traditional username/password authentication with ZITADEL in the ShadowAPI project.

## 🎯 Overview

Instead of using OAuth2 flows, this approach uses direct username/password authentication against ZITADEL's session API.

## 🔧 Configuration

### Current Setup
- **ZITADEL URL**: `http://auth.localtest.me`
- **Admin User**: `admin@example.com` / `Admin123!`
- **Session API**: `http://auth.localtest.me/v2beta/sessions`

## 🛠️ Frontend Implementation

### 1. Authentication Service

Create `front/src/auth/passwordAuthService.ts`:

```typescript
interface LoginCredentials {
  username: string;
  password: string;
}

interface UserSession {
  sessionId: string;
  sessionToken: string;
  userId: string;
  username: string;
}

class PasswordAuthService {
  private readonly STORAGE_KEYS = {
    SESSION_TOKEN: 'auth_session_token',
    USER_INFO: 'auth_user_info'
  };

  private readonly ZITADEL_URL = 'http://auth.localtest.me';

  // Login with username/password
  async login(credentials: LoginCredentials): Promise<boolean> {
    try {
      // Step 1: Create session with username/password
      const response = await fetch(`${this.ZITADEL_URL}/v2beta/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          checks: {
            user: {
              loginName: credentials.username
            },
            password: {
              password: credentials.password
            }
          }
        })
      });

      if (!response.ok) {
        throw new Error('Authentication failed');
      }

      const session = await response.json();

      // Store session information
      localStorage.setItem(this.STORAGE_KEYS.SESSION_TOKEN, session.sessionToken);
      localStorage.setItem(this.STORAGE_KEYS.USER_INFO, JSON.stringify({
        userId: session.factors.user.userId,
        username: credentials.username,
        sessionId: session.sessionId
      }));

      return true;
    } catch (error) {
      console.error('Login failed:', error);
      return false;
    }
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!localStorage.getItem(this.STORAGE_KEYS.SESSION_TOKEN);
  }

  // Get current user
  getCurrentUser(): any {
    const userInfo = localStorage.getItem(this.STORAGE_KEYS.USER_INFO);
    return userInfo ? JSON.parse(userInfo) : null;
  }

  // Get session token for API calls
  getSessionToken(): string | null {
    return localStorage.getItem(this.STORAGE_KEYS.SESSION_TOKEN);
  }

  // Logout
  logout(): void {
    localStorage.removeItem(this.STORAGE_KEYS.SESSION_TOKEN);
    localStorage.removeItem(this.STORAGE_KEYS.USER_INFO);
  }
}

export const passwordAuthService = new PasswordAuthService();
```

### 2. Login Form Component

Create `front/src/components/LoginForm.tsx`:

```typescript
import React, { useState } from 'react';
import { passwordAuthService } from '../auth/passwordAuthService';
import { useNavigate } from 'react-router-dom';

export function LoginForm() {
  const [credentials, setCredentials] = useState({ username: '', password: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const success = await passwordAuthService.login(credentials);

      if (success) {
        navigate('/dashboard');
      } else {
        setError('Invalid username or password');
      }
    } catch (err) {
      setError('Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <form onSubmit={handleSubmit} className="login-form">
        <h2>Login to ShadowAPI</h2>

        {error && (
          <div className="error-message" style={{ color: 'red', marginBottom: '1rem' }}>
            {error}
          </div>
        )}

        <div className="form-group">
          <label htmlFor="username">Email:</label>
          <input
            id="username"
            type="email"
            value={credentials.username}
            onChange={(e) => setCredentials(prev => ({ ...prev, username: e.target.value }))}
            placeholder="admin@example.com"
            required
            disabled={loading}
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Password:</label>
          <input
            id="password"
            type="password"
            value={credentials.password}
            onChange={(e) => setCredentials(prev => ({ ...prev, password: e.target.value }))}
            placeholder="Password"
            required
            disabled={loading}
          />
        </div>

        <button type="submit" disabled={loading} className="login-button">
          {loading ? 'Signing in...' : 'Sign In'}
        </button>
      </form>

      <div className="demo-credentials" style={{ marginTop: '1rem', fontSize: '0.9rem', color: '#666' }}>
        <p><strong>Demo Credentials:</strong></p>
        <p>Email: admin@example.com</p>
        <p>Password: Admin123!</p>
      </div>
    </div>
  );
}
```

### 3. Protected Route Component

Create `front/src/components/ProtectedRoute.tsx`:

```typescript
import React from 'react';
import { Navigate } from 'react-router-dom';
import { passwordAuthService } from '../auth/passwordAuthService';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  if (!passwordAuthService.isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
```

### 4. API Integration

For API calls to your backend, include the session token:

```typescript
// In your API service
import { passwordAuthService } from '../auth/passwordAuthService';

export async function apiCall(url: string, options: RequestInit = {}) {
  const sessionToken = passwordAuthService.getSessionToken();

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
    ...(sessionToken && { 'X-Session-Token': sessionToken })
  };

  return fetch(url, { ...options, headers });
}
```

## 🎨 Basic CSS Styling

Add to your CSS file:

```css
.login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
}

.login-form {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

.login-form h2 {
  text-align: center;
  margin-bottom: 1.5rem;
  color: #333;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #555;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.form-group input:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.login-button {
  width: 100%;
  padding: 0.75rem;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.login-button:hover:not(:disabled) {
  background-color: #0056b3;
}

.login-button:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.error-message {
  padding: 0.75rem;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  color: #721c24;
}

.demo-credentials {
  background-color: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  border: 1px solid #e9ecef;
}
```

## 🔧 Backend Integration

Update your backend to handle session tokens instead of OAuth tokens:

```go
// In your backend auth middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sessionToken := r.Header.Get("X-Session-Token")
        if sessionToken == "" {
            http.Error(w, "Missing session token", http.StatusUnauthorized)
            return
        }

        // Validate session token with ZITADEL
        // Add user context to request
        next.ServeHTTP(w, r)
    })
}
```

## 🎯 Usage

### Default Credentials
- **Username**: `admin@example.com`
- **Password**: `Admin123!`

### Login Flow
1. User enters username/password
2. Frontend calls ZITADEL session API
3. ZITADEL validates credentials and returns session
4. Session token stored locally
5. Token included in API calls to backend

## 🔄 Clean Setup Commands

```bash
# Clean up old PKCE setup
rm -f zitadel-pkce-config.json secrets/.zitadel-setup-complete

# Remove PKCE setup service (optional)
# Edit compose.yaml to remove zitadel-setup service
```

## ✅ Benefits

✅ **Simple**: Traditional username/password flow
✅ **Familiar**: Standard form-based authentication
✅ **Direct**: No redirects or complex OAuth flows
✅ **Custom UI**: Full control over login experience
✅ **Session-based**: Uses ZITADEL's session management

This approach gives you a traditional login form while still leveraging ZITADEL for user management and authentication! 🎯