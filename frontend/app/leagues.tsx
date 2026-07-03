import { useQuery } from "@tanstack/react-query";
import { getLeagues } from "~/lib/api";
import { Card } from "~/components/ui/Card";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import type { LeagueInfo } from "~/lib/types";

export default function LeaguesPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["leagues"],
    queryFn: () => getLeagues(),
    staleTime: 24 * 60 * 60 * 1000,
  });

  if (isError) return <div className="max-w-4xl mx-auto py-12 px-4"><ErrorCard onRetry={() => refetch()} /></div>;
  if (isLoading) return (
    <div className="max-w-4xl mx-auto py-14 px-4 space-y-4">
      <Skeleton className="h-8 w-40" />
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        {Array.from({ length: 12 }).map((_, i) => <Skeleton key={i} className="h-36 rounded-2xl" />)}
      </div>
    </div>
  );

  return (
    <div className="max-w-4xl mx-auto py-14 px-4 space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-text-primary">联赛段位</h1>
        <p className="text-sm text-text-secondary mt-1.5">Clash of Clans 所有联赛等级</p>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-5">
        {(data?.items ?? []).map((league: LeagueInfo) => (
          <Card key={league.id} className="card-hover shadow-card text-center p-6">
            {league.iconUrls?.small ? (
              <img src={league.iconUrls.small} alt="" className="w-14 h-14 mx-auto mb-4 object-contain drop-shadow-sm" />
            ) : (
              <div className="w-14 h-14 mx-auto mb-4 rounded-xl bg-primary-bg flex items-center justify-center text-text-muted text-2xl">?</div>
            )}
            <p className="text-sm font-semibold text-text-primary">{league.name}</p>
          </Card>
        ))}
      </div>
    </div>
  );
}
