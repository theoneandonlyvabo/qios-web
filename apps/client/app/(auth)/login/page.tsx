"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Eye, EyeOff, Loader2, AlertCircle, CheckCircle2, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

export default function LoginPage() {
  const [showPassword, setShowPassword] = React.useState(false);
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [isSuccess, setIsSuccess] = React.useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    
    // Simulate loading for UI state demonstration
    setTimeout(() => {
      setIsLoading(false);
      // For demo purposes, we can toggle between error and success
      const isDemoSuccess = false; // Change to true to see success state
      
      if (isDemoSuccess) {
        setIsSuccess(true);
      } else {
        setError("Kredensial yang Anda masukkan tidak cocok dengan data kami.");
      }
    }, 2000);
  };

  return (
    <div className="relative min-h-screen w-full flex items-center justify-center bg-background overflow-hidden p-4 sm:p-6">
      {/* Premium Background Elements */}
      <div className="absolute inset-0 z-0">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-brand/5 blur-[120px] rounded-full animate-pulse" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-brand/10 blur-[120px] rounded-full animate-pulse" style={{ animationDelay: '2s' }} />
        <div className="absolute inset-0 bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] dark:bg-[radial-gradient(#ffffff0a_1px,transparent_1px)] [background-size:32px_32px] [mask-image:radial-gradient(ellipse_50%_50%_at_50%_50%,#000_70%,transparent_100%)]" />
      </div>

      {/* Top Navigation */}
      <nav className="absolute top-0 left-0 right-0 p-6 flex justify-between items-center z-20">
        <div className="flex items-center gap-2 group cursor-default">
          <div className="h-8 w-8 bg-brand rounded-lg flex items-center justify-center text-white font-bold text-lg shadow-lg shadow-brand/20 group-hover:scale-105 transition-transform">
            Q
          </div>
          <span className="font-bold text-xl tracking-tight text-foreground">QIOS</span>
        </div>
        <ThemeToggle />
      </nav>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="relative z-10 w-full max-w-[440px]"
      >
        <Card className="glass border-border/50 shadow-2xl overflow-hidden">
          <CardHeader className="space-y-2 pb-6">
            <motion.div
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.2 }}
            >
              <CardTitle className="text-3xl font-extrabold tracking-tight">Selamat Datang</CardTitle>
              <CardDescription className="text-base">
                Silakan masuk untuk mengelola operasional bisnis Anda.
              </CardDescription>
            </motion.div>
          </CardHeader>

          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="space-y-2">
                <label
                  htmlFor="email"
                  className="text-sm font-medium text-foreground/80 ml-1"
                >
                  Alamat Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="name@business.com"
                  required
                  autoComplete="email"
                  className="bg-muted/30 border-border/50 focus:bg-background transition-colors"
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between ml-1">
                  <label
                    htmlFor="password"
                    className="text-sm font-medium text-foreground/80"
                  >
                    Kata Sandi
                  </label>
                </div>
                <div className="relative group">
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••"
                    required
                    autoComplete="current-password"
                    className="pr-12 bg-muted/30 border-border/50 focus:bg-background transition-colors"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-0 top-0 h-full w-12 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>

              <AnimatePresence mode="wait">
                {error && (
                  <motion.div
                    initial={{ opacity: 0, height: 0, y: -10 }}
                    animate={{ opacity: 1, height: "auto", y: 0 }}
                    exit={{ opacity: 0, height: 0, y: -10 }}
                    className="flex items-start gap-3 rounded-xl bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20"
                  >
                    <AlertCircle className="h-5 w-5 shrink-0 mt-0.5" />
                    <p className="leading-relaxed font-medium">{error}</p>
                  </motion.div>
                )}

                {isSuccess && (
                  <motion.div
                    initial={{ opacity: 0, height: 0, y: -10 }}
                    animate={{ opacity: 1, height: "auto", y: 0 }}
                    exit={{ opacity: 0, height: 0, y: -10 }}
                    className="flex items-start gap-3 rounded-xl bg-green-500/10 p-4 text-sm text-green-600 dark:text-green-400 border border-green-500/20"
                  >
                    <CheckCircle2 className="h-5 w-5 shrink-0 mt-0.5" />
                    <p className="leading-relaxed font-medium">Berhasil! Mengalihkan ke dashboard...</p>
                  </motion.div>
                )}
              </AnimatePresence>

              <Button
                type="submit"
                className="w-full relative group overflow-hidden"
                disabled={isLoading || isSuccess}
              >
                <span className={cn(
                  "flex items-center justify-center gap-2 transition-all",
                  isLoading && "opacity-0"
                )}>
                  Masuk Sekarang
                  <ArrowRight size={18} className="group-hover:translate-x-1 transition-transform" />
                </span>
                
                {isLoading && (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <Loader2 className="h-5 w-5 animate-spin" />
                  </div>
                )}
              </Button>
            </form>
          </CardContent>

          <CardFooter className="flex flex-col space-y-4 pt-4 pb-8 border-t border-border/30 bg-muted/10">
            <p className="text-center text-sm text-muted-foreground px-4">
              Akses terbatas untuk partner QIOS. 
              <br className="hidden sm:block" />
              Hubungi administrator untuk mendapatkan akun.
            </p>
          </CardFooter>
        </Card>
        
        <motion.div 
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.6 }}
          className="mt-8 flex flex-col items-center gap-4"
        >
          <div className="flex items-center gap-6">
            <span className="text-xs font-medium text-muted-foreground/60 uppercase tracking-widest">Powered by</span>
            <div className="h-px w-8 bg-border" />
            <span className="text-xs font-bold text-muted-foreground/80 tracking-widest">SKALAR SOLUTIONS</span>
          </div>
          <p className="text-[10px] text-muted-foreground/40 font-medium">
            &copy; 2026 QIOS SYSTEM. SEMUA HAK DILINDUNGI.
          </p>
        </motion.div>
      </motion.div>
    </div>
  );
}
