'use client';

import { useEffect, useState, useCallback } from 'react';
import { authApi } from '../lib/api';

// ==========================================
// useAuth — placeholder token management
// Access token disimpan di memory (intentional, sesuai CLAUDE.md)
// Refresh token ada di HttpOnly cookie (diset server)
// ==========================================

interface AuthState {
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

let inMemoryToken: string | null = null;

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    accessToken: null,
    isAuthenticated: false,
    isLoading: true,
  });

  // Rehydrate token dari localStorage saat app load (offline mode support)
  useEffect(() => {
    const stored = typeof window !== 'undefined'
      ? localStorage.getItem('qios_access_token')
      : null;

    if (stored) {
      inMemoryToken = stored;
      setState({ accessToken: stored, isAuthenticated: true, isLoading: false });
    } else {
      setState((prev) => ({ ...prev, isLoading: false }));
    }
  }, []);

  const setToken = useCallback((token: string) => {
    inMemoryToken = token;
    localStorage.setItem('qios_access_token', token);
    setState({ accessToken: token, isAuthenticated: true, isLoading: false });
  }, []);

  const clearToken = useCallback(() => {
    inMemoryToken = null;
    localStorage.removeItem('qios_access_token');
    setState({ accessToken: null, isAuthenticated: false, isLoading: false });
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch {
      // Tetap clear token meski request gagal
    } finally {
      clearToken();
    }
  }, [clearToken]);

  return {
    ...state,
    setToken,
    clearToken,
    logout,
    /** Token untuk dipakai di Authorization header */
    getToken: () => inMemoryToken,
  };
}
