import type { ReactNode } from "react";

interface CardProps {
  children: ReactNode;
  className?: string;
}
interface CompoundProps {
  children: ReactNode;
  className?: string;
}

export function Card({ children, className = "" }: CardProps) {
  return (
    <div className={`bg-surface-card rounded-card border border-border-light ${className}`}
      style={{ boxShadow: "0 1px 4px rgba(30,33,48,0.04), 0 4px 16px rgba(30,33,48,0.04)" }}>
      {children}
    </div>
  );
}

Card.Header = function CardHeader({ children, className = "" }: CompoundProps) {
  return <div className={`px-6 pt-5 pb-0 ${className}`}>{children}</div>;
};

Card.Body = function CardBody({ children, className = "" }: CompoundProps) {
  return <div className={`px-6 py-5 ${className}`}>{children}</div>;
};

Card.Footer = function CardFooter({ children, className = "" }: CompoundProps) {
  return (
    <div className={`px-6 pb-5 flex gap-3 border-t border-border-light pt-4 ${className}`}>
      {children}
    </div>
  );
};
