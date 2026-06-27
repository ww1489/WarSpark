import type { ReactNode } from "react";

const colors: Record<string, string> = {
  green: "bg-emerald-600 text-white",
  red: "bg-rose-600 text-white",
  yellow: "bg-amber-500 text-white",
  blue: "bg-blue-600 text-white",
  gray: "bg-slate-600 text-white",
  primary: "bg-primary text-white",
  purple: "bg-purple-600 text-white",
};

export function Badge({ color = "gray", children }: { color?: string; children: ReactNode }) {
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-tag text-xs font-medium ${colors[color] ?? colors.gray}`}>
      {children}
    </span>
  );
}
