import { Button } from "./Button";

interface ErrorCardProps {
  title?: string;
  message?: string;
  onRetry?: () => void;
}

export function ErrorCard({ title = "加载失败", message = "请稍后重试", onRetry }: ErrorCardProps) {
  return (
    <div className="bg-surface-card rounded-card p-12 text-center space-y-4">
      <div className="text-4xl">😞</div>
      <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
      <p className="text-text-secondary text-sm">{message}</p>
      {onRetry && (
        <Button variant="secondary" size="sm" onClick={onRetry}>
          重新加载
        </Button>
      )}
    </div>
  );
}
