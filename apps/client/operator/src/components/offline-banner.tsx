"use client";

import { useEffect, useState } from "react";
import { Icon } from "@/components/icons";

export function OfflineBanner() {
  const [online, setOnline] = useState(true);

  useEffect(() => {
    setOnline(window.navigator.onLine);
    const handleOnline = () => setOnline(true);
    const handleOffline = () => setOnline(false);
    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);
    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  if (online) return null;

  return (
    <div className="mx-4 mt-3 flex items-center gap-2 rounded-xl border border-warning/25 bg-warning/10 px-3 py-2 text-xs font-bold text-warning">
      <Icon name="wifi" className="h-4 w-4" />
      Offline — data terakhir tetap bisa dilihat. Checkout dinonaktifkan sampai koneksi balik.
    </div>
  );
}
