import * as React from "react"
import { cn } from "@/lib/utils"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'secondary'
  size?: 'default' | 'sm' | 'lg' | 'icon'
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', ...props }, ref) => {
    const baseStyles = "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-semibold transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]"
    
    const variants = {
      default: "bg-brand text-white shadow-[0_1px_2px_rgba(0,0,0,0.1),0_4px_12px_rgba(204,34,0,0.2)] hover:bg-brand-soft hover:shadow-[0_1px_2px_rgba(0,0,0,0.1),0_8px_20px_rgba(204,34,0,0.3)]",
      outline: "border border-border bg-transparent shadow-sm hover:bg-accent hover:text-accent-foreground",
      secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
      ghost: "hover:bg-accent hover:text-accent-foreground"
    }
    
    const sizes = {
      default: "h-11 px-6 py-2",
      sm: "h-9 px-4 text-xs",
      lg: "h-13 px-10 text-base rounded-xl",
      icon: "h-10 w-10"
    }

    return (
      <button
        ref={ref}
        className={cn(baseStyles, variants[variant], sizes[size], className)}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
