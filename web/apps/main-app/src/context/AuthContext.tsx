import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { client, setClientAccessToken } from '@graphfolio/api-client';

export interface AuthUser {
  userId: string;
  email: string;
  displayName: string;
  displayCurrency: string;
  theme: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AuthContextType {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string, rememberMe?: boolean) => Promise<void>;
  register: (email: string, password: string, displayName: string, displayCurrency: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const USER_PREF_SELECTION = {
  userId: true,
  email: true,
  displayName: true,
  displayCurrency: true,
  theme: true,
  createdAt: true,
  updatedAt: true,
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearRefreshTimer = useCallback(() => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  }, []);

  const scheduleSilentRefresh = useCallback((expiresInSeconds: number) => {
    clearRefreshTimer();
    // Refresh 1 minute before expiry, or at least in 30 seconds
    const delayMs = Math.max(30, expiresInSeconds - 60) * 1000;
    refreshTimerRef.current = setTimeout(async () => {
      try {
        const res = await client.mutation({
          refreshToken: {
            accessToken: true,
            expiresIn: true,
            user: USER_PREF_SELECTION,
          },
        });
        if (res.refreshToken) {
          setClientAccessToken(res.refreshToken.accessToken);
          if (res.refreshToken.user) {
            setUser(res.refreshToken.user as AuthUser);
          }
          scheduleSilentRefresh(res.refreshToken.expiresIn);
        }
      } catch (err) {
        console.warn('Silent token refresh failed:', err);
        setClientAccessToken(null);
        setUser(null);
      }
    }, delayMs);
  }, [clearRefreshTimer]);

  // Initial silent refresh attempt on app mount (using existing httpOnly cookie)
  useEffect(() => {
    let isMounted = true;

    async function initAuth() {
      try {
        const res = await client.mutation({
          refreshToken: {
            accessToken: true,
            expiresIn: true,
            user: USER_PREF_SELECTION,
          },
        });

        if (isMounted && res.refreshToken) {
          setClientAccessToken(res.refreshToken.accessToken);
          if (res.refreshToken.user) {
            setUser(res.refreshToken.user as AuthUser);
          }
          scheduleSilentRefresh(res.refreshToken.expiresIn);
        }
      } catch (err) {
        // Unauthenticated or no refresh cookie present
        if (isMounted) {
          setClientAccessToken(null);
          setUser(null);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    initAuth();

    return () => {
      isMounted = false;
      clearRefreshTimer();
    };
  }, [scheduleSilentRefresh, clearRefreshTimer]);

  const login = useCallback(
    async (email: string, password: string, rememberMe: boolean = false) => {
      setIsLoading(true);
      setError(null);
      try {
        const res = await client.mutation({
          login: {
            __args: {
              input: {
                email,
                password,
                rememberMe,
              },
            },
            accessToken: true,
            expiresIn: true,
            user: USER_PREF_SELECTION,
          },
        });

        if (!res.login) {
          throw new Error('Login failed: invalid response');
        }

        setClientAccessToken(res.login.accessToken);
        if (res.login.user) {
          setUser(res.login.user as AuthUser);
        }
        scheduleSilentRefresh(res.login.expiresIn);
      } catch (err: any) {
        const msg = err.message || 'Login failed. Please check your credentials.';
        setError(msg);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [scheduleSilentRefresh]
  );

  const register = useCallback(
    async (email: string, password: string, displayName: string, displayCurrency: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const res = await client.mutation({
          register: {
            __args: {
              input: {
                email,
                password,
                displayName,
                displayCurrency,
              },
            },
            accessToken: true,
            expiresIn: true,
            user: USER_PREF_SELECTION,
          },
        });

        if (!res.register) {
          throw new Error('Registration failed: invalid response');
        }

        setClientAccessToken(res.register.accessToken);
        if (res.register.user) {
          setUser(res.register.user as AuthUser);
        }
        scheduleSilentRefresh(res.register.expiresIn);
      } catch (err: any) {
        const msg = err.message || 'Registration failed. Please try again.';
        setError(msg);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [scheduleSilentRefresh]
  );

  const logout = useCallback(async () => {
    setIsLoading(true);
    clearRefreshTimer();
    try {
      await client.mutation({
        logout: true,
      });
    } catch (err) {
      console.warn('Logout mutation error:', err);
    } finally {
      setClientAccessToken(null);
      setUser(null);
      setIsLoading(false);
    }
  }, [clearRefreshTimer]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value = {
    user,
    isAuthenticated: !!user,
    isLoading,
    error,
    login,
    register,
    logout,
    clearError,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
