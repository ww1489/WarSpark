import { useParams } from "@tanstack/react-router";
import { useState, useCallback, useMemo } from "react";
import { useClan, useWarLog, useRaidSeasons, useAllLeagues } from "~/lib/queries";
import { getWarLog } from "~/lib/api";
import { Card } from "~/components/ui/Card";
import { Tabs } from "~/components/ui/Tabs";
import { Badge } from "~/components/ui/Badge";
import { Table } from "~/components/ui/Table";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import { formatNumber, formatDate, roleLabel, normalizeTag } from "~/lib/utils";
import type { ClanMemberSummary, WarLogEntry, CapitalRaidSeason } from "~/lib/types";

const TABS = [
  { key: "overview", label: "成员" },
  { key: "warLog", label: "战争日志" },
  { key: "raids", label: "突袭赛季" },
];

export default function ClanDetailPage() {
  const { tag } = useParams({ strict: false }) as { tag: string };
  const normalized = normalizeTag(tag);
  const { data, isLoading, isError, refetch } = useClan(normalized);
  const { data: leagueData } = useAllLeagues();
  const [activeTab, setActiveTab] = useState("overview");

  const leagueIcons = useMemo(() => {
    const map: Record<number, string> = {};
    for (const l of leagueData?.items ?? []) {
      if (l.iconUrls?.small) map[l.id] = l.iconUrls.small;
    }
    return map;
  }, [leagueData]);

  if (isError) return <div className="max-w-7xl mx-auto p-12"><ErrorCard onRetry={() => refetch()} /></div>;

  return (
    <div className="max-w-7xl mx-auto py-8 px-6 space-y-6">
      {isLoading ? <Skeleton.Card /> : data ? <ClanBanner clan={data.clan} /> : null}
      <Card>
        <Tabs tabs={TABS} active={activeTab} onChange={setActiveTab} />
        {activeTab === "overview" && <OverviewTab data={data} isLoading={isLoading} leagueIcons={leagueIcons} />}
        {activeTab === "warLog" && <WarLogTab tag={normalized} />}
        {activeTab === "raids" && <RaidsTab tag={normalized} />}
      </Card>
    </div>
  );
}

function ClanBanner({ clan }: { clan: NonNullable<ReturnType<typeof useClan>["data"]>["clan"] }) {
  return (
    <Card>
      <Card.Body className="flex flex-wrap items-start gap-5">
        <img src={clan.badgeUrls.large} alt="" className="w-24 h-24 rounded-2xl object-contain bg-white border border-border-light" />
        <div className="flex-1 min-w-0 space-y-2">
          <h1 className="text-2xl font-bold tracking-tight text-text-primary">{clan.name}</h1>
          <p className="text-sm text-text-secondary font-mono">{clan.tag}</p>
          <div className="flex gap-2 flex-wrap">
            <Badge color="primary">Lv.{clan.clanLevel}</Badge>
            {clan.warLeague && <Badge color="purple">{clan.warLeague.name}</Badge>}
            {clan.type === "inviteOnly" ? <Badge color="yellow">仅邀请</Badge> : clan.type === "open" ? <Badge color="green">开放</Badge> : <Badge color="red">关闭</Badge>}
          </div>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MiniStat label="成员" value={String(clan.members)} />
          <MiniStat label="积分" value={formatNumber(clan.clanPoints)} />
          <MiniStat label="战绩" value={`${clan.warWins}胜`} sub={`${clan.warWinStreak ?? 0}连胜`} />
          <MiniStat label="战争频率" value={clan.warFrequency === "always" ? "长期" : clan.warFrequency ?? "-"} />
        </div>
      </Card.Body>
      {clan.description && (
        <Card.Body className="pt-0">
          <p className="text-sm text-text-secondary leading-relaxed bg-surface-page rounded-xl p-4">{clan.description}</p>
        </Card.Body>
      )}
      {clan.labels && clan.labels.length > 0 && (
        <Card.Body className="pt-0">
          <div className="flex gap-2 flex-wrap">
            {clan.labels.map((l) => (
              <span key={l.id} className="px-2.5 py-1 text-xs bg-primary-bg text-primary rounded-full">{l.name}</span>
            ))}
          </div>
        </Card.Body>
      )}
    </Card>
  );
}

function MiniStat({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="bg-surface-page rounded-xl px-4 py-3 text-center min-w-[80px]">
      <div className="text-xs text-text-muted mb-0.5">{label}</div>
      <div className="text-lg font-bold text-text-primary tabular-nums">{value}</div>
      {sub && <div className="text-xs text-text-secondary mt-0.5">{sub}</div>}
    </div>
  );
}

function OverviewTab({ data, isLoading, leagueIcons }: {
  data: ReturnType<typeof useClan>["data"];
  isLoading: boolean;
  leagueIcons: Record<number, string>;
}) {
  if (isLoading) return <Skeleton.Table rows={10} cols={7} />;
  if (!data) return null;

  return (
    <div>
      <div className="flex items-center gap-2 px-4 py-2 border-b border-border text-xs font-medium text-text-muted">
        <span className="w-6 text-center">#</span>
        <span className="w-6" />
        <span className="flex-1">成员</span>
        <span className="w-16 text-center">角色</span>
        <span className="w-10 text-center">等级</span>
        <span className="w-10 text-center">大本</span>
        <span className="w-16 text-right">奖杯</span>
        <span className="w-20 text-right">捐/收</span>
      </div>
      {data.members.map((row, i) => {
        const icon = row.leagueTier?.iconUrls?.small || (row.league?.id ? leagueIcons[row.league.id] : null);
        return (
          <div key={i} className="flex items-center gap-2 px-4 py-2 border-b border-border-light">
            <span className="w-6 text-xs text-text-muted">{row.clanRank}</span>
            {icon ? <img src={icon} alt="" className="w-6 h-6" /> : <span className="w-6 h-6 block" />}
            <div className="min-w-0 flex-1">
              <div className="text-sm font-medium truncate">{row.name}</div>
              <div className="text-xs text-text-muted truncate">{row.tag}</div>
            </div>
            <span className="text-xs w-16 text-center">{roleLabel(row.role)}</span>
            <span className="text-sm w-10 text-center text-text-secondary">{row.expLevel}</span>
            <span className="text-sm w-10 text-center text-primary font-medium">TH{row.townHallLevel}</span>
            <span className="text-sm w-16 text-right tabular-nums">{formatNumber(row.trophies)}</span>
            <span className="text-sm w-20 text-right tabular-nums">
              <span className="text-text-primary">{formatNumber(row.donations)}</span>
              <span className="text-text-muted">/{formatNumber(row.donationsReceived ?? 0)}</span>
            </span>
          </div>
        );
      })}
    </div>
  );
}

function WarLogTab({ tag }: { tag: string }) {
  const { data, isLoading, isError, error, refetch } = useWarLog(tag, 15);
  const [loadedPages, setLoadedPages] = useState<{ items: WarLogEntry[]; cursor?: string; more: boolean }[]>([]);
  const [page, setPage] = useState(0);
  const [fetching, setFetching] = useState(false);

  const page0items = data?.items ?? [];
  const page0cursor = data?.paging?.cursors?.after;
  const page0more = !!page0cursor;

  const currentItems = page === 0 ? page0items : (loadedPages[page - 1]?.items ?? []);
  const lastMore = loadedPages.length > 0 ? loadedPages[loadedPages.length - 1]?.more : page0more;
  const totalPages = 1 + loadedPages.length + (lastMore ? 1 : 0);

  const goTo = useCallback(async (idx: number) => {
    if (idx < 0) return;
    setPage(idx);
    if (idx === 0) return;
    const slot = idx - 1;
    if (slot < loadedPages.length || fetching) return;
    setFetching(true);
    try {
      const pages = [...loadedPages];
      while (pages.length <= slot) {
        const lastCursor = pages.length === 0 ? page0cursor : pages[pages.length - 1]?.cursor;
        if (!lastCursor) break;
        const resp = await getWarLog(tag, 15, lastCursor);
        pages.push({
          items: resp.items ?? [],
          cursor: resp.paging?.cursors?.after,
          more: !!resp.paging?.cursors?.after,
        });
      }
      setLoadedPages(pages);
    } finally {
      setFetching(false);
    }
  }, [tag, loadedPages, fetching, page0cursor]);

  const errStatus = (error as { status?: number })?.status;

  if (isError) return (
    <div className="py-12 text-center">
      <p className="text-text-secondary text-sm">
        {errStatus === 403 ? "该部落未公开战争日志" : "获取失败，请稍后重试"}
      </p>
      {errStatus !== 403 && (
        <button onClick={() => refetch()} className="text-xs text-primary hover:underline mt-3">重试</button>
      )}
    </div>
  );

  if (isLoading && currentItems.length === 0) return <div className="py-12"><Skeleton.Table rows={5} cols={6} /></div>;
  if (currentItems.length === 0 && !isLoading) return <div className="py-10 text-center text-text-muted text-sm">暂无战争日志</div>;

  return (
    <div className="overflow-hidden">
      <div>
        {/* header */}
        <div className="grid items-center py-2.5" style={{ gridTemplateColumns: "100px 1fr 60px 60px 72px 64px" }}>
          <div className="px-4 table-header text-center">日期</div>
          <div className="px-4 table-header text-center">对手</div>
          <div className="px-4 table-header text-center">规模</div>
          <div className="px-4 table-header text-center">结果</div>
          <div className="px-4 table-header text-center">星数</div>
          <div className="px-4 table-header text-center">摧毁</div>
        </div>
        {/* body */}
        <div className="divide-y divide-border-light">
          {currentItems.map((row, i) => {
            const resultColor = row.result === "win" ? "bg-emerald-100" : row.result === "lose" ? "bg-rose-100" : "bg-slate-100";
            const badCls = row.result === "win" ? "bg-emerald-600 text-white" : row.result === "lose" ? "bg-rose-600 text-white" : "bg-slate-600 text-white";
            const txt = row.result === "win" ? "胜利" : row.result === "lose" ? "失败" : "平局";

            return (
              <div
                key={i}
                className={`grid items-center py-3 transition-colors duration-150 ${resultColor}`}
                style={{ gridTemplateColumns: "100px 1fr 60px 60px 72px 64px" }}
              >
                <div className="px-4 text-sm tabular-nums truncate text-center">{formatDate(row.endTime)}</div>
                <div className="px-4 text-sm tabular-nums truncate text-center">
                  <span className="text-left block truncate">{row.opponent?.name || "未知部落"}</span>
                </div>
                <div className="px-4 text-sm tabular-nums truncate text-center">{row.teamSize}v{row.teamSize}</div>
                <div className="px-4 text-sm tabular-nums truncate text-center">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-tag text-xs font-medium ${badCls}`}>{txt}</span>
                </div>
                <div className="px-4 text-sm tabular-nums truncate text-center">{row.clan?.stars ?? 0} - {row.opponent?.stars ?? 0}</div>
                <div className="px-4 text-sm tabular-nums truncate text-center">{row.clan?.destructionPercentage?.toFixed(1) ?? "0"}%</div>
              </div>
            );
          })}
        </div>
        {/* pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 py-4">
            <button
              onClick={() => goTo(page - 1)}
              disabled={page === 0 || fetching}
              className="px-3 py-1.5 text-sm rounded-btn border border-border text-text-secondary hover:bg-surface-page disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              上一页
            </button>
            {Array.from({ length: Math.min(totalPages, 10) }, (_, i) => (
              <button
                key={i}
                onClick={() => goTo(i)}
                disabled={fetching}
                className={`w-8 h-8 text-sm rounded-btn transition-colors ${
                  i === page
                    ? "bg-primary text-white"
                    : "text-text-secondary hover:bg-surface-page"
                }`}
              >
                {i + 1}
              </button>
            ))}
            <button
              onClick={() => goTo(page + 1)}
              disabled={page >= totalPages - 1 || fetching}
              className="px-3 py-1.5 text-sm rounded-btn border border-border text-text-secondary hover:bg-surface-page disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              下一页
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function RaidsTab({ tag }: { tag: string }) {
  const { data, isLoading, isError, refetch } = useRaidSeasons(tag);
  const [page, setPage] = useState(0);
  const PER_PAGE = 10;

  const allItems = data?.items ?? [];
  const totalPages = Math.ceil(allItems.length / PER_PAGE);
  const pageItems = allItems.slice(page * PER_PAGE, (page + 1) * PER_PAGE);

  if (isError) return <div className="p-6"><ErrorCard onRetry={() => refetch()} /></div>;
  if (isLoading) return <div className="p-6"><Skeleton.Table rows={3} cols={5} /></div>;
  if (!allItems.length) return <div className="p-6 text-center text-text-secondary text-sm">暂无突袭数据</div>;

  return (
    <div className="p-6 space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        {pageItems.map((season: CapitalRaidSeason, i: number) => (
          <div key={i} className="bg-surface-page rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${season.state === "ongoing" ? "bg-emerald-50 text-emerald-700" : "bg-gray-100 text-gray-500"}`}>
                {season.state === "ongoing" ? "进行中" : "已结束"}
              </span>
              <span className="text-xs text-text-muted">{formatDate(season.startTime)} - {formatDate(season.endTime)}</span>
            </div>
            <div className="grid grid-cols-3 gap-3 text-center">
              <div>
                <div className="text-2xl font-bold tabular-nums text-text-primary">{season.raidsCompleted}</div>
                <div className="text-xs text-text-muted mt-1">突袭完成</div>
              </div>
              <div>
                <div className="text-2xl font-bold tabular-nums text-text-primary">{season.totalAttacks}</div>
                <div className="text-xs text-text-muted mt-1">总攻击</div>
              </div>
              <div>
                <div className="text-2xl font-bold tabular-nums text-text-primary">{formatNumber(season.capitalTotalLoot)}</div>
                <div className="text-xs text-text-muted mt-1">总战利品</div>
              </div>
            </div>
          </div>
        ))}
      </div>
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-2">
          <button onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page === 0}
            className="px-3 py-1.5 text-sm rounded-btn border border-border text-text-secondary hover:bg-surface-page disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            上一页
          </button>
          {Array.from({ length: Math.min(totalPages, 10) }, (_, i) => (
            <button key={i} onClick={() => setPage(i)}
              className={`w-8 h-8 text-sm rounded-btn transition-colors ${i === page ? "bg-primary text-white" : "text-text-secondary hover:bg-surface-page"}`}>
              {i + 1}
            </button>
          ))}
          <button onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))} disabled={page >= totalPages - 1}
            className="px-3 py-1.5 text-sm rounded-btn border border-border text-text-secondary hover:bg-surface-page disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            下一页
          </button>
        </div>
      )}
    </div>
  );
}
