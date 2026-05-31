'use client';

import { useEffect, useState, useCallback } from 'react';
import { operatorApi } from '@/lib/api';

const ACCESS_TOKEN_KEY = 'operator_access_token';

const getToken = () => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
};

const setToken = (token: string) => {
  if (typeof window === 'undefined') return;
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
};

const clearToken = () => {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(ACCESS_TOKEN_KEY);
};

interface AuthState {
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    accessToken: null,
    isAuthenticated: false,
    isLoading: true,
  });

  useEffect(() => {
    const token = getToken();
    setState({
      accessToken: token,
      isAuthenticated: !!token,
      isLoading: false,
    });
  }, []);

  const loginQr = useCallback(async (qrToken: string) => {
    const res = await operatorApi.loginQr(qrToken);
    setToken(res.access_token);
    setState({ accessToken: res.access_token, isAuthenticated: true, isLoading: false });
  }, []);

  const loginCredential = useCallback(async (body: {
    business_id: string;
    operator_code: string;
    password: string;
  }) => {
    const res = await operatorApi.loginCredential(body);
    setToken(res.access_token);
    setState({ accessToken: res.access_token, isAuthenticated: true, isLoading: false });
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setState({ accessToken: null, isAuthenticated: false, isLoading: false });
  }, []);

  return { ...state, loginQr, loginCredential, logout };
}