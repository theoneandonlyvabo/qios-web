"use client";

import type { ReactNode } from "react";
import { OperatorProvider } from "@/components/providers/operator-provider";

export function AppProviders({ children }: { children: ReactNode }) {
  return <OperatorProvider>{children}</OperatorProvider>;
}
