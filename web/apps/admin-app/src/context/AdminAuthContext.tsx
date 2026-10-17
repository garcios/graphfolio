import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { client, setClientAccessToken } from '@graphfolio/api-client';

export interface AuthUser {
  userId: string;
  email: string;
  displayName: string;
  displayCurrency: string;
  theme: string;
  role: 'USER' | 'ADMIN';
  mustChangePassword: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface AdminAuthContextType {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  changePassword: (currentPassword: string, newPassword: string, revokeOtherSessions?: boolean) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

const AdminAuthContext = createContext<AdminAuthContextType | undefined>(undefined);

const USER_PREF_SELECTION = {
  userId: true,
  email: true,
  displayName: true,
  displayCurrency: true,
  theme: true,
  role: true,
  mustChangePassword: true,
  createdAt: true,
  updatedAt: true,
};

export const AdminAuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
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

  // Initial silent refresh attempt on mount
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
        // No active refresh token cookie or session expired
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

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await client.mutation({
        login: {
          __args: {
            input: {
              email: email.trim().toLowerCase(),
              password,
              rememberMe: false,
            },
          },
          accessToken: true,
          expiresIn: true,
          user: USER_PREF_SELECTION,
        },
      });

      if (!res.login) {
        throw new Error('Authentication failed');
      }

      setClientAccessToken(res.login.accessToken);
      if (res.login.user) {
        setUser(res.login.user as AuthUser);
      }
      scheduleSilentRefresh(res.login.expiresIn);
    } catch (err: any) {
      const msg = err.message || 'Login failed. Please verify your credentials.';
      setError(msg);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [scheduleSilentRefresh]);

  const changePassword = useCallback(async (
    currentPassword: string,
    newPassword: string,
    revokeOtherSessions: boolean = true
  ) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await client.mutation({
        changePassword: {
          __args: {
            input: {
              currentPassword,
              newPassword,
              revokeOtherSessions,
            },
          },
          success: true,
          message: true,
        },
      });

      if (!res.changePassword?.success) {
        throw new Error(res.changePassword?.message || 'Failed to change password');
      }

      // Mark mustChangePassword as false locally
      setUser(prev => prev ? { ...prev, mustChangePassword: false } : null);
    } catch (err: any) {
      const msg = err.message || 'Failed to change password';
      setError(msg);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await client.mutation({
        logout: true,
      });
    } catch (err) {
      console.warn('Logout error (clearing local session anyway):', err);
    } finally {
      clearRefreshTimer();
      setClientAccessToken(null);
      setUser(null);
      setError(null);
      setIsLoading(false);
    }
  }, [clearRefreshTimer]);

  const value: AdminAuthContextType = {
    user,
    isAuthenticated: Boolean(user),
    isLoading,
    error,
    login,
    changePassword,
    logout,
    clearError,
  };

  return <AdminAuthContext.Provider value={value}>{children}</AdminAuthContext.Provider>;
};

export const useAdminAuth = (): AdminAuthContextType => {
  const context = useContext(AdminAuthContext);
  if (!context) {
    throw new Error('useAdminAuth must be used within an AdminAuthProvider');
  }
  return context;
};
