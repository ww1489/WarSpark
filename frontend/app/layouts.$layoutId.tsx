import { useState } from "react";
import { useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { getLayout } from "~/lib/api";
import { Card } from "~/components/ui/Card";
import { Badge } from "~/components/ui/Badge";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import type { LayoutImage, LayoutCard, VideoMatch } from "~/lib/types";

const layoutTypeLabel: Record<string, string> = {
  war: "部落战", farming: "种田", trophy: "冲杯", hybrid: "混合",
};

export default function LayoutDetailPage() {
  const { layoutId } = useParams({ strict: false }) as { layoutId: string };

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["layout", layoutId],
    queryFn: () => getLayout(layoutId),
    enabled: !!layoutId,
  });

  if (isError) return <div className="max-w-5xl mx-auto py-12 px-4"><ErrorCard onRetry={() => refetch()} /></div>;
  if (isLoading) return <div className="max-w-5xl mx-auto py-12 px-4 space-y-4"><Skeleton className="h-[400px] rounded-2xl" /><Skeleton className="h-48 rounded-2xl" /></div>;
  if (!data) return null;

  return (
    <div className="max-w-5xl mx-auto py-10 px-4 space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-text-primary">{data.title}</h1>
        <div className="flex gap-2 mt-3">
          <Badge color="primary">TH{data.th_level}</Badge>
          <Badge color="gray">{layoutTypeLabel[data.layout_type] ?? data.layout_type}</Badge>
          {data.style_tags?.map((tag) => <Badge key={tag} color="purple">{tag}</Badge>)}
        </div>
      </div>

      {data.images.length > 0 && <Gallery images={data.images} />}

      {data.links.length > 0 && (
        <Card className="shadow-card">
          <div className="px-5 py-4 border-b border-border">
            <h3 className="font-semibold text-text-primary">阵型链接</h3>
          </div>
          <div className="p-3 space-y-2">
            {data.links.map((link) => (
              <a
                key={link.link_id}
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-3 p-3 bg-surface-page rounded-xl hover:bg-primary-bg/30 transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-primary-bg flex items-center justify-center text-xs shrink-0">🔗</div>
                <span className="text-xs text-primary truncate flex-1 font-mono">{link.url}</span>
                <Badge color="gray">{link.link_type}</Badge>
              </a>
            ))}
          </div>
        </Card>
      )}

      {data.attack_videos.length > 0 && <VideoSection title="进攻视频" videos={data.attack_videos} />}
      {data.defense_replays.length > 0 && <VideoSection title="防守回放" videos={data.defense_replays} />}

      {data.similar_layouts.length > 0 && <SimilarGrid layouts={data.similar_layouts} />}
    </div>
  );
}

function Gallery({ images }: { images: LayoutImage[] }) {
  const [active, setActive] = useState(0);

  return (
    <Card className="overflow-hidden shadow-card">
      <div className="aspect-[16/10] bg-surface-page flex items-center justify-center p-2">
        <img src={images[active].image_url} alt="" className="max-w-full max-h-full object-contain rounded-lg" />
      </div>
      {images.length > 1 && (
        <div className="flex gap-2 p-3 overflow-x-auto border-t border-border">
          {images.map((img, i) => (
            <button
              key={img.image_id}
              type="button"
              onClick={() => setActive(i)}
              className={`shrink-0 w-20 h-14 rounded-lg overflow-hidden transition-all ${
                i === active ? "ring-2 ring-primary scale-105 shadow-md" : "opacity-50 hover:opacity-80 hover:scale-105"
              }`}
            >
              <img src={img.image_url} alt="" className="w-full h-full object-cover" />
            </button>
          ))}
        </div>
      )}
    </Card>
  );
}

function VideoSection({ title, videos }: { title: string; videos: VideoMatch[] }) {
  return (
    <Card className="shadow-card overflow-hidden">
      <div className="px-5 py-4 border-b border-border">
        <h3 className="font-semibold text-text-primary">{title}</h3>
      </div>
      <div className="p-3 space-y-2">
        {videos.map((v) => (
          <a
            key={v.match_id}
            href={v.youtube_url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-4 p-3 bg-surface-page rounded-xl hover:bg-primary-bg/30 transition-colors group"
          >
            <div className="w-10 h-10 rounded-xl bg-primary-bg flex items-center justify-center text-sm shrink-0 group-hover:bg-primary group-hover:text-white transition-colors">▶</div>
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
    </Card>
  );
}

function SimilarGrid({ layouts }: { layouts: LayoutCard[] }) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-text-primary">相似阵型</h3>
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
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
              <div className="p-3">
                <p className="text-xs font-semibold text-text-primary truncate group-hover:text-primary transition-colors mb-1.5">{l.title}</p>
                <Badge color="primary">TH{l.th_level}</Badge>
              </div>
            </Card>
          </a>
        ))}
      </div>
    </div>
  );
}
