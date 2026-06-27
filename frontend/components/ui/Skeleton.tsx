export function Skeleton({ className = "" }: { className?: string }) {
  return <div className={`animate-pulse bg-gray-200 rounded ${className}`} />;
}

Skeleton.Card = function SkeletonCard() {
  return (
    <div className="bg-surface-card rounded-card p-6 space-y-4">
      <Skeleton className="h-6 w-1/3" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-2/3" />
    </div>
  );
};

Skeleton.Table = function SkeletonTable({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="space-y-3 p-6">
      {Array.from({ length: rows }).map((_, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: skeleton placeholders
        <div key={i} className="flex gap-4">
          {Array.from({ length: cols }).map((_, j) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: skeleton placeholders
            <Skeleton key={j} className={`h-4 ${j === 0 ? "w-12" : "flex-1"}`} />
          ))}
        </div>
      ))}
    </div>
  );
};
