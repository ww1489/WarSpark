import { useMemo, useState } from "react";
import { useLocations, useRankings, type RankingType } from "~/lib/queries";
import { Card } from "~/components/ui/Card";
import { Tabs } from "~/components/ui/Tabs";
import { Table } from "~/components/ui/Table";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import { BatchAvatarSelect } from "~/components/ui/BatchAvatarSelect";
import { formatNumber, rankChangeText } from "~/lib/utils";
import type {
  ClanRankingEntry, PlayerRankingEntry, ClanCapitalEntry,
  ClanBuilderBaseEntry, PlayerBuilderBaseEntry,
} from "~/lib/types";

const DEFAULT_LOCATION = "32000017";

const TABS: { key: RankingType; label: string }[] = [
  { key: "clans", label: "部落" },
  { key: "players", label: "玩家" },
  { key: "capital", label: "都城" },
  { key: "builder-clans", label: "夜世界·部落" },
  { key: "builder-players", label: "夜世界·玩家" },
];

export default function LeaderboardsPage() {
  const { data: locations, isLoading: locsLoading } = useLocations();
  const [locationId, setLocationId] = useState(DEFAULT_LOCATION);
  const [activeTab, setActiveTab] = useState<RankingType>("clans");

  const locationOptions = useMemo(() => {
    if (!locations?.items) return [];
    return locations.items.map((l) => ({ value: String(l.id), label: l.name }));
  }, [locations]);

  return (
    <div className="max-w-7xl mx-auto py-8 px-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-text-primary">排行榜</h1>
        <BatchAvatarSelect value={locationId} onChange={setLocationId} options={locationOptions} loading={locsLoading} />
      </div>
      <Card>
        <Tabs tabs={TABS} active={activeTab} onChange={setActiveTab} />
        <RankingTable type={activeTab} locationId={locationId} />
      </Card>
    </div>
  );
}

function RankingTable({ type, locationId }: { type: RankingType; locationId: string }) {
  const { data, isLoading, isError, refetch } = useRankings(locationId, type);

  if (isError) return <div className="p-6"><ErrorCard onRetry={() => refetch()} /></div>;
  if (isLoading) return <div className="p-6"><Skeleton.Table rows={12} cols={6} /></div>;
  if (!data?.items.length) return <div className="p-8 text-center text-text-muted text-sm">该地区暂无排行数据</div>;

  switch (type) {
    case "clans": return <ClanTable data={data.items as ClanRankingEntry[]} />;
    case "players": return <PlayerTable data={data.items as PlayerRankingEntry[]} />;
    case "capital": return <CapitalTable data={data.items as ClanCapitalEntry[]} />;
    case "builder-clans": return <BuilderClanTable data={data.items as ClanBuilderBaseEntry[]} />;
    case "builder-players": return <BuilderPlayerTable data={data.items as PlayerBuilderBaseEntry[]} />;
  }
}

function RankCell({ rank, prev }: { rank: number; prev: number }) {
  const cls = rank === 1 ? "rank-1" : rank === 2 ? "rank-2" : rank === 3 ? "rank-3" : "bg-gray-200 text-gray-500";
  return (
    <div className="flex items-center justify-center gap-1.5">
      <span className={`inline-flex w-6 h-6 rounded-full items-center justify-center text-xs font-bold text-white ${cls}`}>
        {rank}
      </span>
      {rankChangeText(rank, prev) && (
        <span className="text-xs text-text-muted tabular-nums w-6">{rankChangeText(rank, prev)}</span>
      )}
    </div>
  );
}

function ClanBadge({ url, name }: { url: string; name: string }) {
  return <img src={url} alt={name} className="w-8 h-8 rounded-lg object-contain" />;
}

function NameLink({ tag, name }: { tag: string; name: string }) {
  return (
    <a href={`/clans/${tag.replace("#", "%23")}`} className="text-left block truncate text-primary hover:underline font-medium">
      {name}
    </a>
  );
}

function PlayerLink({ tag, name }: { tag: string; name: string }) {
  return (
    <a href={`/players/${tag.replace("#", "%23")}`} className="text-left block truncate text-primary hover:underline font-medium">
      {name}
    </a>
  );
}

// ── Clan Rankings ──
function ClanTable({ data }: { data: ClanRankingEntry[] }) {
  return (
    <Table
      columns={[
        { key: "rank", header: "排名", width: "80px", render: (r: ClanRankingEntry) => <RankCell rank={r.rank} prev={r.previousRank} /> },
        { key: "badge", header: "", width: "52px", render: (r: ClanRankingEntry) => <ClanBadge url={r.badgeUrls.small} name={r.name} /> },
        { key: "name", header: "部落名称", width: "1fr", render: (r: ClanRankingEntry) => <NameLink tag={r.tag} name={r.name} /> },
        { key: "level", header: "等级", width: "52px", render: (r: ClanRankingEntry) => <span className="text-text-secondary">{r.clanLevel}</span> },
        { key: "points", header: "积分", width: "80px", render: (r: ClanRankingEntry) => <span className="tabular-nums font-medium">{formatNumber(r.clanPoints)}</span> },
        { key: "members", header: "成员", width: "52px", render: (r: ClanRankingEntry) => <span className="text-text-secondary">{r.members}</span> },
        { key: "location", header: "地区", width: "100px", render: (r: ClanRankingEntry) => <span className="text-text-muted text-xs">{r.location?.name ?? "-"}</span> },
      ]}
      data={data}
    />
  );
}

// ── Player Rankings ──
function PlayerTable({ data }: { data: PlayerRankingEntry[] }) {
  return (
    <Table
      columns={[
        { key: "rank", header: "排名", width: "80px", render: (r: PlayerRankingEntry) => <RankCell rank={r.rank} prev={r.previousRank} /> },
        { key: "name", header: "玩家名称", width: "1fr", render: (r: PlayerRankingEntry) => (
          <div className="text-left flex items-center gap-2">
            {r.leagueTier?.iconUrls?.small ? (
              <img src={r.leagueTier.iconUrls.small} alt="" className="w-5 h-5 object-contain shrink-0" />
            ) : null}
            <PlayerLink tag={r.tag} name={r.name} />
          </div>
        )},
        { key: "level", header: "等级", width: "52px", render: (r: PlayerRankingEntry) => <span className="text-text-secondary">{r.expLevel}</span> },
        { key: "trophies", header: "奖杯", width: "76px", render: (r: PlayerRankingEntry) => <span className="tabular-nums font-medium">{formatNumber(r.trophies)}</span> },
        { key: "attackWins", header: "进攻胜场", width: "76px", render: (r: PlayerRankingEntry) => <span className="text-text-secondary">{formatNumber(r.attackWins ?? 0)}</span> },
        { key: "defenseWins", header: "防御胜场", width: "76px", render: (r: PlayerRankingEntry) => <span className="text-text-secondary">{formatNumber(r.defenseWins ?? 0)}</span> },
        { key: "clan", header: "部落", width: "120px", render: (r: PlayerRankingEntry) => r.clan?.name ? (
          <a href={`/clans/${r.clan.tag.replace("#", "%23")}`} className="text-text-secondary hover:text-primary text-xs block truncate text-left">
            {r.clan.name}
          </a>
        ) : <span className="text-text-muted text-xs">-</span> },
      ]}
      data={data}
    />
  );
}

// ── Capital Rankings ──
function CapitalTable({ data }: { data: ClanCapitalEntry[] }) {
  return (
    <Table
      columns={[
        { key: "rank", header: "排名", width: "80px", render: (r: ClanCapitalEntry) => <RankCell rank={r.rank} prev={r.previousRank} /> },
        { key: "badge", header: "", width: "52px", render: (r: ClanCapitalEntry) => <ClanBadge url={r.badgeUrls.small} name={r.name} /> },
        { key: "name", header: "部落名称", width: "1fr", render: (r: ClanCapitalEntry) => <NameLink tag={r.tag} name={r.name} /> },
        { key: "level", header: "等级", width: "52px", render: (r: ClanCapitalEntry) => <span className="text-text-secondary">{r.clanLevel}</span> },
        { key: "capitalPoints", header: "都城积分", width: "88px", render: (r: ClanCapitalEntry) => <span className="tabular-nums font-medium">{formatNumber(r.clanCapitalPoints)}</span> },
        { key: "members", header: "成员", width: "52px", render: (r: ClanCapitalEntry) => <span className="text-text-secondary">{r.members}</span> },
        { key: "location", header: "地区", width: "100px", render: (r: ClanCapitalEntry) => <span className="text-text-muted text-xs">{r.location?.name ?? "-"}</span> },
      ]}
      data={data}
    />
  );
}

// ── Builder Clan Rankings ──
function BuilderClanTable({ data }: { data: ClanBuilderBaseEntry[] }) {
  return (
    <Table
      columns={[
        { key: "rank", header: "排名", width: "80px", render: (r: ClanBuilderBaseEntry) => <RankCell rank={r.rank} prev={r.previousRank} /> },
        { key: "badge", header: "", width: "52px", render: (r: ClanBuilderBaseEntry) => <ClanBadge url={r.badgeUrls.small} name={r.name} /> },
        { key: "name", header: "部落名称", width: "1fr", render: (r: ClanBuilderBaseEntry) => <NameLink tag={r.tag} name={r.name} /> },
        { key: "level", header: "等级", width: "52px", render: (r: ClanBuilderBaseEntry) => <span className="text-text-secondary">{r.clanLevel}</span> },
        { key: "bbPoints", header: "夜世界积分", width: "96px", render: (r: ClanBuilderBaseEntry) => <span className="tabular-nums font-medium">{formatNumber(r.clanBuilderBasePoints)}</span> },
        { key: "members", header: "成员", width: "52px", render: (r: ClanBuilderBaseEntry) => <span className="text-text-secondary">{r.members}</span> },
        { key: "location", header: "地区", width: "100px", render: (r: ClanBuilderBaseEntry) => <span className="text-text-muted text-xs">{r.location?.name ?? "-"}</span> },
      ]}
      data={data}
    />
  );
}

// ── Builder Player Rankings ──
function BuilderPlayerTable({ data }: { data: PlayerBuilderBaseEntry[] }) {
  return (
    <Table
      columns={[
        { key: "rank", header: "排名", width: "80px", render: (r: PlayerBuilderBaseEntry) => <RankCell rank={r.rank} prev={r.previousRank} /> },
        { key: "name", header: "玩家名称", width: "1fr", render: (r: PlayerBuilderBaseEntry) => <PlayerLink tag={r.tag} name={r.name} /> },
        { key: "level", header: "等级", width: "52px", render: (r: PlayerBuilderBaseEntry) => <span className="text-text-secondary">{r.expLevel}</span> },
        { key: "trophies", header: "夜世界奖杯", width: "96px", render: (r: PlayerBuilderBaseEntry) => <span className="tabular-nums font-medium">{formatNumber(r.builderBaseTrophies)}</span> },
        { key: "clan", header: "部落", width: "120px", render: (r: PlayerBuilderBaseEntry) => r.clan?.name ? (
          <a href={`/clans/${r.clan.tag.replace("#", "%23")}`} className="text-text-secondary hover:text-primary text-xs block truncate text-left">
            {r.clan.name}
          </a>
        ) : <span className="text-text-muted text-xs">-</span> },
      ]}
      data={data}
    />
  );
}
