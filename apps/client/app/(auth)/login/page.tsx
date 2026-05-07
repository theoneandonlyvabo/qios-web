"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Eye, 
  EyeOff, 
  Loader2, 
  AlertCircle, 
  LayoutDashboard, 
  ShieldCheck, 
  Zap,
  ChevronRight
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";

const features = [
  {
    icon: LayoutDashboard,
    title: "Dashboard Terintegrasi",
    description: "Pantau semua data finansial bisnis dalam satu tampilan intuitif."
  },
  {
    icon: Zap,
    title: "Otomasi Transaksi",
    description: "Hemat waktu dengan pencatatan otomatis dan rekonsiliasi instan."
  },
  {
    icon: ShieldCheck,
    title: "Keamanan Enterprise",
    description: "Data bisnis Anda terlindungi dengan enkripsi tingkat perbankan."
  }
];

export default function LoginPage() {
  const [showPassword, setShowPassword] = React.useState(false);
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    
    setTimeout(() => {
      setIsLoading(false);
      setError("Email atau kata sandi yang Anda masukkan salah.");
    }, 2000);
  };

  return (
    <div className="relative min-h-screen w-full flex flex-col lg:flex-row bg-background overflow-x-hidden">
      {/* Background Decorative Elements */}
      <div className="absolute inset-0 z-0 pointer-events-none overflow-hidden">
        <div className="absolute top-[-10%] left-[-5%] w-[50%] h-[50%] bg-brand/5 blur-[120px] rounded-full" />
        <div className="absolute bottom-[-10%] right-[-5%] w-[50%] h-[50%] bg-brand-orange/5 blur-[120px] rounded-full" />
        <div className="absolute inset-0 bg-dot-grid opacity-[0.3] dark:opacity-[0.1]" />
      </div>

      {/* Theme Toggle Floating */}
      <div className="absolute top-4 right-4 lg:top-6 lg:right-6 z-50">
        <ThemeToggle />
      </div>

      {/* LEFT SIDE: Branding & Marketing */}
      <section className="relative z-10 w-full lg:w-1/2 flex flex-col justify-between p-6 sm:p-10 lg:p-12 xl:p-20 border-b lg:border-b-0 lg:border-r border-border/50">
        <div className="flex items-center gap-2.5 mb-6 lg:mb-12">
          <div className="h-8 w-8 lg:h-10 lg:w-10 bg-gradient-to-br from-brand to-brand-orange rounded-xl flex items-center justify-center text-white font-bold text-base lg:text-xl shadow-lg shadow-brand/20">
            Q
          </div>
          <span className="font-bold text-lg lg:text-2xl tracking-tight text-foreground">QIOS</span>
        </div>

        <div className="max-w-xl flex-1 flex flex-col justify-center py-8 lg:pt-12 lg:pb-24">
          <motion.h1 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="text-2xl sm:text-4xl lg:text-5xl xl:text-6xl font-extrabold tracking-tight leading-[1.2] lg:leading-[1.1] mb-4 lg:mb-6"
          >
            Kendali Penuh Atas <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand to-brand-orange">
              Data Bisnis Anda.
            </span>
          </motion.h1>
          
          <motion.p 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="text-sm sm:text-base lg:text-lg text-muted-foreground mb-0 lg:mb-12 leading-relaxed max-w-md lg:max-w-none"
          >
            Satu sistem terintegrasi untuk manajemen finansial dan operasional.
            Dirancang khusus untuk UMKM yang siap naik kelas.
          </motion.p>

          <div className="hidden lg:block space-y-6 xl:space-y-8 mt-12">
            {features.map((feature, idx) => (
              <motion.div 
                key={idx}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.5, delay: 0.2 + (idx * 0.1) }}
                className="flex items-start gap-4 group"
              >
                <div className="mt-1 h-10 w-10 shrink-0 rounded-lg bg-brand-dim flex items-center justify-center text-brand group-hover:bg-brand group-hover:text-white transition-all duration-300">
                  <feature.icon size={20} />
                </div>
                <div>
                  <h3 className="font-bold text-foreground mb-1">{feature.title}</h3>
                  <p className="text-sm text-muted-foreground leading-snug">{feature.description}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        <div className="mt-16 lg:mt-auto flex items-center gap-4 text-xs font-medium text-muted-foreground/50 pt-8">
          <span>&copy; 2026 QIOS SYSTEM</span>
          <div className="h-3 w-px bg-border/50" />
          <span>POWERED BY SKALAR</span>
        </div>
      </section>

      {/* RIGHT SIDE: Login Form */}
      <section className="relative z-10 w-full lg:w-1/2 flex items-center justify-center p-4 sm:p-8 lg:p-12 xl:p-24 bg-muted/10 lg:bg-transparent">
        {/* Glow effect behind card on mobile */}
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-brand/[0.03] to-transparent lg:hidden" />
        
        <motion.div
          initial={{ opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
          className="w-full max-w-[440px] relative"
        >
          {/* Accent glow for the card in dark mode */}
          <div className="absolute -inset-2 bg-gradient-to-br from-brand/10 to-brand-orange/10 rounded-[2.5rem] blur-3xl opacity-0 dark:opacity-30 transition-opacity" />

          <Card className="glass relative border-border/40 shadow-xl rounded-[1.5rem] lg:rounded-[2rem] overflow-hidden premium-shadow flex flex-col">
            <CardHeader className="space-y-1 p-6 lg:p-10 lg:pb-6">
              <div className="lg:hidden flex items-center gap-2 mb-4">
                <div className="h-7 w-7 bg-brand rounded-lg flex items-center justify-center text-white font-bold shadow-md">Q</div>
                <span className="font-bold text-lg tracking-tight">QIOS</span>
              </div>
              <CardTitle className="text-2xl lg:text-3xl font-extrabold tracking-tight">Masuk Akun</CardTitle>
              <CardDescription className="text-sm lg:text-base text-muted-foreground/80">
                Akses portal manajemen QIOS untuk bisnis Anda.
              </CardDescription>
            </CardHeader>

            <CardContent className="p-6 lg:p-10 pt-0 lg:pt-0">
              <form onSubmit={handleSubmit} className="space-y-4 lg:space-y-5">
                <div className="space-y-2">
                  <label htmlFor="email" className="text-[13px] font-semibold text-foreground/70 ml-1">
                    Email Bisnis
                  </label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="admin@bisnis-anda.com"
                    required
                    className="bg-muted/30 border-border/40 h-11 lg:h-12 rounded-xl px-4 focus:bg-card focus:ring-1 focus:ring-brand/20 transition-all"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between ml-1">
                    <label htmlFor="password" className="text-[13px] font-semibold text-foreground/70">
                      Kata Sandi
                    </label>
                    <button type="button" className="text-[12px] font-bold text-brand hover:text-brand-orange transition-colors">
                      Lupa sandi?
                    </button>
                  </div>
                  <div className="relative group">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="••••••••"
                      required
                      className="bg-muted/30 border-border/40 h-11 lg:h-12 rounded-xl pl-4 pr-12 focus:bg-card focus:ring-1 focus:ring-brand/20 transition-all"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-0 top-0 h-full w-12 flex items-center justify-center text-muted-foreground/60 hover:text-brand transition-colors"
                    >
                      {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                </div>

                <AnimatePresence mode="wait">
                  {error && (
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: "auto" }}
                      exit={{ opacity: 0, height: 0 }}
                      className="overflow-hidden"
                    >
                      <div className="flex items-center gap-2 rounded-xl bg-destructive/5 p-3 text-xs lg:text-[13px] text-destructive border border-destructive/10 mt-1">
                        <AlertCircle className="h-4 w-4 shrink-0" />
                        <p className="font-medium leading-tight">{error}</p>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>

                <Button
                  type="submit"
                  variant="brand"
                  className="w-full h-11 lg:h-12 text-sm lg:text-[15px] font-bold flex items-center justify-center gap-2 group mt-2"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    <>
                      Masuk Sekarang
                      <ChevronRight size={18} className="group-hover:translate-x-1 transition-transform" />
                    </>
                  )}
                </Button>
              </form>
            </CardContent>

            <CardFooter className="flex flex-col space-y-4 p-6 lg:p-10 pt-0 lg:pt-0 bg-muted/[0.02]">
              <div className="w-full flex items-center gap-4 py-2">
                <div className="h-px flex-1 bg-border/40" />
                <span className="text-[10px] font-bold text-muted-foreground/40 uppercase tracking-[0.2em]">Atau</span>
                <div className="h-px flex-1 bg-border/40" />
              </div>

              <div className="grid grid-cols-2 gap-3 w-full">
                <Button variant="outline" className="rounded-xl h-10 lg:h-11 text-xs font-bold gap-2 border-border/40 hover:bg-muted/50 transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24">
                    <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                    <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                    <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" />
                    <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                  </svg>
                  Google
                </Button>
                <Button variant="outline" className="rounded-xl h-10 lg:h-11 text-xs font-bold gap-2 border-border/40 hover:bg-muted/50 transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M17.05 20.28c-.96.95-2.04 1.72-3.24 2.31c-1.2.59-2.48.88-3.84.88c-1.36 0-2.64-.29-3.84-.88c-1.2-.59-2.28-1.36-3.24-2.31c-.96-.95-1.72-2.04-2.31-3.24c-.59-1.2-.88-2.48-.88-3.84c0-1.36.29-2.64.88-3.84c.59-1.2 1.35-2.28 2.31-3.24c.96-.95 2.04-1.72 3.24-2.31c1.2-.59 2.48-.88 3.84-.88c1.36 0 2.64.29 3.84.88c1.2.59 2.28 1.36 3.24 2.31l-2.04 2.04c-.65-.65-1.37-1.16-2.16-1.53c-.79-.37-1.63-.56-2.52-.56c-.89 0-1.73.19-2.52.56c-.79.37-1.51.88-2.16 1.53c-.65.65-1.16 1.37-1.53 2.16c-.37.79-.56 1.63-.56 2.52s.19 1.73.56 2.52c.37.79.88 1.51 1.53 2.16c.65.65 1.37 1.16 2.16 1.53c.79.37 1.63.56 2.52.56c.89 0 1.73-.19 2.52-.56c.79-.37 1.51-.88 2.16-1.53l2.04 2.04z" />
                  </svg>
                  Apple
                </Button>
              </div>

              <p className="text-center text-xs text-muted-foreground/60">
                Belum terdaftar? <button className="text-brand font-bold hover:text-brand-orange transition-colors">Hubungi Admin</button>
              </p>
            </CardFooter>
          </Card>
        </motion.div>
      </section>
    </div>
  );
}
