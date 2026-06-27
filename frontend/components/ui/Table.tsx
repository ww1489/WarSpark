import { EmptyState } from "./EmptyState";
import { Skeleton } from "./Skeleton";

interface Column<T> {
  key: string;
  header: string;
  width?: string;
  render?: (row: T) => React.ReactNode;
}

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  emptyText?: string;
  rowKey?: (row: T, index: number) => string;
  onRowClick?: (row: T) => void;
}

export function Table<T>({ columns, data, loading, emptyText, rowKey, onRowClick }: TableProps<T>) {
  if (loading) return <Skeleton.Table rows={5} cols={columns.length} />;
  if (data.length === 0) return <EmptyState title={emptyText ?? "暂无数据"} />;

  const gridCols = columns.map((c) => c.width ?? "1fr").join(" ");

  return (
    <div>
      {/* header */}
      <div
        className="grid items-center py-2.5"
        style={{ gridTemplateColumns: gridCols }}
      >
        {columns.map((col) => (
          <div key={col.key} className="px-4 table-header text-center">
            {col.header}
          </div>
        ))}
      </div>
      {/* body */}
      <div className="divide-y divide-border-light">
        {data.map((row, i) => (
          <div
            key={rowKey ? rowKey(row, i) : `row-${i}`}
            className={`grid items-center py-3 px-0 transition-colors duration-150
              ${i % 2 === 0 ? "row-even" : ""}
              ${onRowClick ? "cursor-pointer row-hover" : "row-hover"}`}
            style={{ gridTemplateColumns: gridCols }}
            tabIndex={onRowClick ? 0 : undefined}
            onClick={() => onRowClick?.(row)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onRowClick?.(row);
              }
            }}
          >
            {columns.map((col) => (
              <div key={col.key} className="px-4 text-sm tabular-nums truncate text-center">
                {col.render ? col.render(row) : String((row as Record<string, unknown>)[col.key] ?? "")}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
