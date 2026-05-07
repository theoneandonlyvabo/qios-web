"use client";

import * as React from "react";
import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { motion, AnimatePresence } from "framer-motion";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="w-[68px] h-9 rounded-full bg-muted/20 animate-pulse border border-border/10" />
    );
  }

  const isDark = theme === "dark";

  return (
    <button
      onClick={() => setTheme(isDark ? "light" : "dark")}
      className="group relative flex h-9 w-[68px] items-center rounded-full bg-muted/40 p-1 transition-colors hover:bg-muted/60 border border-border/20 shadow-sm backdrop-blur-md overflow-hidden"
      aria-label="Toggle theme"
    >
      {/* Background Track Accent - subtle glow in dark mode */}
      <AnimatePresence>
        {isDark && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="absolute inset-0 bg-brand/5 pointer-events-none"
          />
        )}
      </AnimatePresence>

      {/* Sliding Thumb */}
      <motion.div
        className="absolute h-7 w-7 rounded-full bg-background shadow-md border border-border/20 flex items-center justify-center z-10"
        initial={false}
        animate={{
          x: isDark ? 28 : 0,
        }}
        transition={{
          type: "spring",
          stiffness: 400,
          damping: 30,
        }}
      >
        <motion.div
          animate={{ rotate: isDark ? 0 : 0 }}
          transition={{ duration: 0.5 }}
        >
          {isDark ? (
            <Moon className="h-3.5 w-3.5 text-brand" fill="currentColor" />
          ) : (
            <Sun className="h-4 w-4 text-brand-orange" fill="currentColor" />
          )}
        </motion.div>
      </motion.div>

      {/* Static Icons */}
      <div className="flex w-full justify-between px-1.5 text-muted-foreground/40">
        <Sun className="h-4 w-4" />
        <Moon className="h-3.5 w-3.5" />
      </div>
    </button>
  );
}
