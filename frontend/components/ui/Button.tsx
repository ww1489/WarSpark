import type { ButtonHTMLAttributes, ReactNode } from "react";

const variants = {
  primary: "bg-primary text-white hover:bg-primary-dark shadow-sm",
  secondary: "bg-white text-primary border border-primary hover:bg-primary/5",
  ghost: "text-text-secondary hover:text-primary hover:bg-surface-page",
  danger: "bg-danger text-white hover:bg-red-700",
} as const;

const sizes = {
  sm: "h-8 px-4 text-sm rounded-tag",
  md: "h-10 px-6 text-sm rounded-btn",
  lg: "h-12 px-8 text-base rounded-btn",
} as const;

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants;
  size?: keyof typeof sizes;
  loading?: boolean;
  children: ReactNode;
}

export function Button({
  variant = "primary",
  size = "md",
  loading,
  children,
  className = "",
  ...props
}: ButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 font-medium transition-colors
        ${variants[variant]} ${sizes[size]}
        ${loading ? "opacity-60 cursor-not-allowed" : "cursor-pointer"}
        ${className}`}
      disabled={loading}
      {...props}
    >
      {loading && <Spinner />}
      {children}
    </button>
  );
}

function Spinner() {
  return (
    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
  );
}
