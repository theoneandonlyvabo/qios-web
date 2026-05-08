"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { getUserData, clearAuthData, User } from "@/lib/auth";

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const data = getUserData();
    setUser(data);
    setIsLoading(false);
  }, []);

  const logout = () => {
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
