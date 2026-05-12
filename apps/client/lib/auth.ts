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
  if (typeof window !== "undefined") {
    // Set cookie for middleware access
    document.cookie = `${AUTH_TOKEN_KEY}=${data.access_token}; path=/; max-age=86400; SameSite=Lax`;
    
    // Set localStorage for persistence and user data
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
  if (typeof window !== "undefined") {
    document.cookie = `${AUTH_TOKEN_KEY}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
    localStorage.removeItem(AUTH_TOKEN_KEY);
    localStorage.removeItem(USER_DATA_KEY);
  }
};

export const isAuthenticated = (): boolean => {
  return !!getAuthToken();
};
