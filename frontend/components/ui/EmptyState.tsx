interface EmptyStateProps {
  icon?: string;
  title?: string;
  description?: string;
}

export function EmptyState({ icon = "📭", title = "暂无数据", description }: EmptyStateProps) {
  return (
    <div className="py-16 text-center space-y-3">
      <div className="text-4xl">{icon}</div>
      <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
      {description && <p className="text-text-secondary text-sm">{description}</p>}
    </div>
  );
}
