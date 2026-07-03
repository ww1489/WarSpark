import { useParams, useNavigate } from "@tanstack/react-router";
import { getImageSearchJob, getImageSearchResults } from "~/lib/api";
import { useQuery } from "@tanstack/react-query";
import { Card } from "~/components/ui/Card";
import { Button } from "~/components/ui/Button";
import { Badge } from "~/components/ui/Badge";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import type { LayoutCard, VideoMatch } from "~/lib/types";

const layoutTypeLabel: Record<string, string> = {
  war: "部落战", farming: "种田", trophy: "冲杯", hybrid: "混合", default: "未知",
};

export default function FindLayoutResultsPage() {
  const { jobId } = useParams({ strict: false }) as { jobId: string };
  const navigate = useNavigate();

  const { data: job, isError } = useQuery({
    queryKey: ["imageSearchJob", jobId],
    queryFn: () => getImageSearchJob(jobId),
    refetchInterval: (q) =>
      q.state.data?.search_status === "pending" || q.state.data?.search_status === "processing" ? 3000 : false,
  });

  const { data: results, isLoading } = useQuery({
    queryKey: ["imageSearchResults", jobId],
    queryFn: () => getImageSearchResults(jobId),
    enabled: job?.search_status === "completed",
  });

  if (isError)
    return (
      <div className="max-w-4xl mx-auto py-12 px-4">
        <ErrorCard onRetry={() => navigate({ to: "/find-layout" })} />
      </div>
    );

  const isProcessing = !job || job.search_status === "pending" || job.search_status === "processing";

  return (
    <div className="max-w-5xl mx-auto py-10 px-4 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-text-primary">搜索结果</h1>
          <p className="text-xs text-text-muted font-mono mt-1">{jobId}</p>
        </div>
        <Button variant="ghost" onClick={() => navigate({ to: "/find-layout" })}>← 重新上传</Button>
      </div>

      {isProcessing && (
        <Card className="shadow-card border-primary/10">
          <Card.Body className="text-center py-16 space-y-5">
            <div className="w-20 h-20 mx-auto rounded-2xl bg-primary-bg flex items-center justify-center">
              <span className="text-4xl animate-pulse">🔍</span>
            </div>
            <div>
              <h3 className="text-lg font-bold text-text-primary">正在分析截图</h3>
              <p className="text-sm text-text-secondary mt-1">识别大本营等级与建筑布局，匹配最优阵型...</p>
            </div>
            {job?.detected_th ? (
              <Badge color="primary">自动检测 TH{job.detected_th}</Badge>
            ) : (
              <Skeleton className="h-6 w-24 mx-auto rounded-full" />
            )}
          </Card.Body>
        </Card>
      )}

      {job?.search_status === "completed" && isLoading && (
        <div className="space-y-4">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-64 w-full rounded-2xl" />
        </div>
      )}

      {job?.search_status === "completed" && results && (
        <div className="space-y-8">
          {/* Summary */}
          {results.match_summary && (
            <div className={`flex items-center gap-4 px-5 py-4 rounded-2xl border ${results.match_summary.low_confidence ? "bg-warning/5 border-warning/20" : "bg-success/5 border-success/20"}`}>
              <div className={`w-10 h-10 rounded-xl flex items-center justify-center text-lg ${results.match_summary.low_confidence ? "bg-warning/20" : "bg-success/20"}`}>
                {results.match_summary.low_confidence ? "⚠" : "✓"}
              </div>
              <div>
                <p className="text-sm font-semibold text-text-primary">
                  {results.match_summary.low_confidence ? "匹配置信度较低" : "匹配成功"}
                </p>
                <p className="text-xs text-text-secondary mt-0.5">
                  共 {results.match_summary.candidate_count} 个候选阵型
                  {results.match_summary.best_match_level && ` · 最佳 ${results.match_summary.best_match_level}`}
                </p>
              </div>
            </div>
          )}

          <LayoutGrid title="匹配阵型" layouts={results.layouts} />

          {results.attack_videos.length > 0 && <VideoList title="进攻视频" videos={results.attack_videos} />}
          {results.defense_replays.length > 0 && <VideoList title="防守回放" videos={results.defense_replays} />}
        </div>
      )}

      {job?.search_status === "failed" && (
        <Card className="shadow-card">
          <Card.Body className="text-center py-16 space-y-4">
            <div className="w-16 h-16 mx-auto rounded-2xl bg-danger/10 flex items-center justify-center text-2xl">✕</div>
            <h3 className="text-lg font-bold text-text-primary">搜索失败</h3>
            <p className="text-sm text-text-secondary">{job.error_message || "未知错误，请重试"}</p>
            <Button variant="ghost" onClick={() => navigate({ to: "/find-layout" })}>重新上传</Button>
          </Card.Body>
        </Card>
      )}
    </div>
  );
}

function LayoutGrid({ title, layouts }: { title: string; layouts: LayoutCard[] }) {
  if (!layouts.length) return null;
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-text-primary">{title}</h3>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
        {layouts.map((l) => (
          <a key={l.layout_id} href={`/layouts/${l.layout_id}`} className="group block">
            <Card className="h-full overflow-hidden shadow-card group-hover:shadow-card-hover transition-all group-hover:-translate-y-1">
              <div className="aspect-[4/3] bg-surface-page overflow-hidden">
                {l.primary_image_url ? (
                  <img src={l.primary_image_url} alt={l.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-text-muted text-xs">暂无预览</div>
                )}
              </div>
              <div className="p-4 space-y-2">
                <p className="text-sm font-semibold text-text-primary truncate group-hover:text-primary transition-colors">{l.title}</p>
                <div className="flex gap-2">
                  <Badge color="primary">TH{l.th_level}</Badge>
                  <Badge color="gray">{layoutTypeLabel[l.layout_type] ?? l.layout_type}</Badge>
                </div>
              </div>
            </Card>
          </a>
        ))}
      </div>
    </div>
  );
}

function VideoList({ title, videos }: { title: string; videos: VideoMatch[] }) {
  if (!videos.length) return null;
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-text-primary">{title}</h3>
      <div className="space-y-2">
        {videos.slice(0, 5).map((v) => (
          <a
            key={v.match_id}
            href={v.youtube_url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-4 p-4 bg-surface-card border border-border rounded-2xl hover:shadow-card transition-all group"
          >
            <div className="w-10 h-10 rounded-xl bg-primary-bg flex items-center justify-center text-lg shrink-0 group-hover:bg-primary group-hover:text-white transition-colors">
              ▶
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-text-primary truncate group-hover:text-primary transition-colors">{v.video_title}</p>
              {v.channel_name && <p className="text-xs text-text-muted mt-0.5">{v.channel_name}</p>}
            </div>
            {v.timestamp_seconds > 0 && (
              <span className="text-xs font-mono text-text-secondary bg-surface-page rounded-lg px-2 py-1 shrink-0">
                {Math.floor(v.timestamp_seconds / 60)}:{String(v.timestamp_seconds % 60).padStart(2, "0")}
              </span>
            )}
          </a>
        ))}
      </div>
    </div>
  );
}
