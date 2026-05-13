import { setAccessToken, clearAccessToken } from "./api";
import { AUTH_TOKEN_KEY, USER_DATA_KEY } from "./constants";

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  business: any;
}

export interface AuthData {
  access_token: string;
  user: User;
}

export const setAuthData = (data: AuthData) => {
  setAccessToken(data.access_token);
  if (typeof window !== "undefined") {
    // Token stored in localStorage for offline-mode support and memory rehydration.
    // HttpOnly cookie is set server-side by the login route handler.
    localStorage.setItem(AUTH_TOKEN_KEY, data.access_token);
    localStorage.setItem(USER_DATA_KEY, JSON.stringify(data.user));
  }
};

export const getAuthToken = (): string | null => {
  if (typeof window !== "undefined") {
    return localStorage.getItem(AUTH_TOKEN_KEY);
  }
  return null;
};

export const getUserData = (): User | null => {
  if (typeof window !== "undefined") {
    const data = localStorage.getItem(USER_DATA_KEY);
    return data ? JSON.parse(data) : null;
  }
  return null;
};

export const clearAuthData = () => {
  clearAccessToken();
  if (typeof window !== "undefined") {
    localStorage.removeItem(AUTH_TOKEN_KEY);
    localStorage.removeItem(USER_DATA_KEY);
  }
};

export const isAuthenticated = (): boolean => {
  return !!getAuthToken();
};
