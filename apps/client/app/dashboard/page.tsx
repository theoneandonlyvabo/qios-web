"use client";

import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";

export default function DashboardPage() {
  const { user, logout, isLoading } = useAuth();

  if (isLoading) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>;
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      {user && (
        <div className="mb-4">
          <p>Welcome, <strong>{user.name}</strong> ({user.email})</p>
          <p>Role: {user.role}</p>
        </div>
      )}
      <Button onClick={logout}>Logout</Button>
    </div>
  );
}
