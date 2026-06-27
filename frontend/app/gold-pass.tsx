import { useCurrentGoldPass } from "~/lib/queries";
import { Card } from "~/components/ui/Card";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import { formatDate, daysRemaining } from "~/lib/utils";

export default function GoldPassPage() {
  const { data, isLoading, isError, refetch } = useCurrentGoldPass();

  if (isError) return <div className="max-w-7xl mx-auto p-12"><ErrorCard onRetry={() => refetch()} /></div>;

  return (
    <div className="max-w-7xl mx-auto py-8 px-6 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight text-text-primary">Gold Pass</h1>
      {isLoading ? <Skeleton.Card /> : data ? (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card>
            <Card.Body className="text-center py-8">
              <div className="text-xs text-text-muted mb-1 uppercase tracking-wider">赛季开始</div>
              <div className="text-2xl font-bold text-text-primary">{formatDate(data.startTime)}</div>
            </Card.Body>
          </Card>
          <Card>
            <Card.Body className="text-center py-8">
              <div className="text-xs text-text-muted mb-1 uppercase tracking-wider">赛季结束</div>
              <div className="text-2xl font-bold text-text-primary">{formatDate(data.endTime)}</div>
            </Card.Body>
          </Card>
          <Card>
            <Card.Body className="text-center py-8">
              <div className="text-xs text-text-muted mb-1 uppercase tracking-wider">剩余时间</div>
              <div className="text-2xl font-bold text-primary">{daysRemaining(data.endTime)} 天</div>
            </Card.Body>
          </Card>
        </div>
      ) : null}
    </div>
  );
}
