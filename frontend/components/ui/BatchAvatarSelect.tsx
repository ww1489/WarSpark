export function BatchAvatarSelect({
  value, onChange, options, loading,
}: {
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
  loading?: boolean;
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="appearance-none bg-white border border-border rounded-lg px-3 py-2 pr-8 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow"
        disabled={loading}
      >
        {loading
          ? <option>加载中...</option>
          : options.map((o) => (<option key={o.value} value={o.value}>{o.label}</option>))
        }
      </select>
      <svg className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-text-secondary pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    </div>
  );
}
