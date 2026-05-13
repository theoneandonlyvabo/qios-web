"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { getUserData, clearAuthData, getAuthToken, User } from "@/lib/auth";
import { setAccessToken } from "@/lib/api";

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const data = getUserData();
    const token = getAuthToken();
    // Rehydrate in-memory token from localStorage on app load so
    // apiFetch doesn't need to refresh on every page reload.
    if (token) setAccessToken(token);
    setUser(data);
    setIsLoading(false);
  }, []);

  const logout = async () => {
    await fetch("/api/auth/logout", { method: "POST" }).catch(() => {});
    clearAuthData();
    setUser(null);
    router.push("/login");
  };

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    logout,
  };
}
