import { useState } from "react";
import { useParams } from "@tanstack/react-router";
import { usePlayer, useBattleLog, useLeagueGroup } from "~/lib/queries";
import { Card } from "~/components/ui/Card";
import { Badge } from "~/components/ui/Badge";
import { Table } from "~/components/ui/Table";
import { Tabs } from "~/components/ui/Tabs";
import { Skeleton } from "~/components/ui/Skeleton";
import { ErrorCard } from "~/components/ui/ErrorCard";
import { formatNumber, formatDateTime, roleLabel, normalizeTag } from "~/lib/utils";
import type {
  BattleLogEntry,
  HeroLevel,
  TroopSpellLevel,
  LeagueGroupMember,
  AchievementProgress,
  PlayerOverview,
} from "~/lib/types";

const PAGE_TABS = [
  { key: "overview", label: "概览" },
  { key: "army", label: "军队" },
  { key: "achievements", label: "成就" },
  { key: "battlelog", label: "战斗日志" },
  { key: "league", label: "联赛" },
];

export default function PlayerDetailPage() {
  const { tag } = useParams({ strict: false }) as { tag: string };
  const normalized = normalizeTag(tag);
  const { data, isLoading, isError, refetch } = usePlayer(normalized);
  const [activeTab, setActiveTab] = useState("overview");

  if (isError) return <div className="max-w-7xl mx-auto p-12"><ErrorCard onRetry={() => refetch()} /></div>;
  if (isLoading) return <div className="max-w-7xl mx-auto py-8 px-6"><Skeleton.Card /></div>;
  if (!data) return null;

  return (
    <div className="max-w-7xl mx-auto py-8 px-6 space-y-6">
      <PlayerBanner player={data} />
      <Tabs tabs={PAGE_TABS} active={activeTab} onChange={setActiveTab} />
      {activeTab === "overview" && <OverviewTab player={data} />}
      {activeTab === "army" && <ArmyTab player={data} />}
      {activeTab === "achievements" && <AchievementsTab achievements={data.achievements ?? []} />}
      {activeTab === "battlelog" && <BattleLogSection tag={normalized} />}
      {activeTab === "league" && <LeagueGroupSection tag={normalized} />}
    </div>
  );
}

function PlayerBanner({ player }: { player: PlayerOverview }) {
  return (
    <Card>
      <Card.Body className="space-y-4">
        <div className="flex items-center gap-3">
          {player.leagueTier?.iconUrls?.small ? (
            <img src={player.leagueTier.iconUrls.small} alt="" className="w-10 h-10 object-contain shrink-0" />
          ) : null}
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{player.name}</h1>
            <p className="text-sm text-text-secondary font-mono">{player.tag}</p>
          </div>
        </div>

        <div className="flex gap-2 flex-wrap">
          <Badge color="primary">TH{player.townHallLevel}{player.townHallWeaponLevel ? ` ⭐${player.townHallWeaponLevel}` : ""}</Badge>
          {player.builderHallLevel ? <Badge color="orange">BH{player.builderHallLevel}</Badge> : null}
          <Badge color="blue">{roleLabel(player.role ?? "member")}</Badge>
          {player.league && <Badge color="purple">{player.league.name}</Badge>}
          {player.warPreference && (
            <Badge color={player.warPreference === "in" ? "green" : "yellow"}>
              {player.warPreference === "in" ? "参战中" : "未参战"}
            </Badge>
          )}
        </div>

        {player.clan && (
          <div className="flex items-center gap-3 p-3 bg-surface-page rounded-xl">
            <img src={player.clan.badgeUrls.small} alt="" className="w-8 h-8 rounded-lg object-contain" />
            <div>
              <div className="text-sm font-medium text-text-primary">{player.clan.name}</div>
              <div className="text-xs text-text-secondary">{player.clan.tag} · Lv.{player.clan.clanLevel}</div>
            </div>
          </div>
        )}
      </Card.Body>
    </Card>
  );
}

function StatCard({ label, value, sub, raw }: { label: string; value: string | number; sub?: string; raw?: boolean }) {
  return (
    <div className="bg-surface-page rounded-xl px-3 py-3 text-center">
      <div className="text-xs text-text-muted mb-1">{label}</div>
      <div className="text-lg font-bold text-text-primary tabular-nums">
        {raw ? value : formatNumber(value as number)}
      </div>
      {sub && <div className="text-xs text-text-secondary mt-0.5">{sub}</div>}
    </div>
  );
}

function OverviewTab({ player }: { player: PlayerOverview }) {
  return (
    <div className="space-y-6">
      <Card>
        <Card.Header><h3 className="font-semibold text-text-primary">家园战斗</h3></Card.Header>
        <Card.Body>
          <div className="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-6 gap-3">
            <StatCard label="奖杯" value={player.trophies} sub={`最佳 ${formatNumber(player.bestTrophies ?? 0)}`} />
            <StatCard label="战争之星" value={player.warStars ?? 0} />
            <StatCard label="进攻胜场" value={player.attackWins ?? 0} />
            <StatCard label="防御胜场" value={player.defenseWins ?? 0} />
            <StatCard label="经验等级" value={player.expLevel} />
            <StatCard label="大本营" value={`TH${player.townHallLevel}`} raw />
          </div>
        </Card.Body>
      </Card>

      <Card>
        <Card.Header><h3 className="font-semibold text-text-primary">部落贡献</h3></Card.Header>
        <Card.Body>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <StatCard label="捐兵" value={player.donations ?? 0} />
            <StatCard label="收兵" value={player.donationsReceived ?? 0} />
            <StatCard label="都城贡献" value={player.clanCapitalContributions ?? 0} />
          </div>
        </Card.Body>
      </Card>

      {(player.builderBaseTrophies != null || player.builderBaseLeague) && (
        <Card>
          <Card.Header><h3 className="font-semibold text-text-primary">夜世界</h3></Card.Header>
          <Card.Body>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <StatCard label="夜世界杯数" value={player.builderBaseTrophies ?? 0} />
              <StatCard label="最佳夜世界" value={player.bestBuilderBaseTrophies ?? 0} />
              {player.builderBaseLeague && <StatCard label="夜本联赛" value={player.builderBaseLeague.name} raw />}
            </div>
          </Card.Body>
        </Card>
      )}

      {player.legendStatistics && (
        <Card>
          <Card.Header><h3 className="font-semibold text-text-primary">传说联赛</h3></Card.Header>
          <Card.Body>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <StatCard label="传说奖杯" value={player.legendStatistics.legendTrophies ?? 0} />
              {player.legendStatistics.currentSeason && (
                <>
                  <StatCard label="当前排名" value={`#${(player.legendStatistics.currentSeason.rank ?? 0).toLocaleString()}`} raw />
                  <StatCard label="当前杯数" value={player.legendStatistics.currentSeason.trophies ?? 0} />
                </>
              )}
              {player.legendStatistics.bestSeason && (
                <StatCard label="最佳赛季杯数" value={player.legendStatistics.bestSeason.trophies ?? 0} sub={`排名 #${(player.legendStatistics.bestSeason.rank ?? 0).toLocaleString()}`} />
              )}
            </div>
          </Card.Body>
        </Card>
      )}

      {player.labels && player.labels.length > 0 && (
        <div className="flex gap-2 flex-wrap">
          {player.labels.map((l) => (
            <span key={l.id} className="px-2.5 py-1 text-xs bg-primary-bg text-primary rounded-full">{l.name}</span>
          ))}
        </div>
      )}
    </div>
  );
}

function ArmyTab({ player }: { player: PlayerOverview }) {
  const [activeSub, setActiveSub] = useState("heroes");
  const subTabs = [
    { key: "heroes", label: "英雄" },
    ...(player.heroEquipment?.length ? [{ key: "equipment", label: "装备" }] : []),
    ...(player.troops?.length ? [{ key: "troops", label: "兵种" }] : []),
    ...(player.spells?.length ? [{ key: "spells", label: "法术" }] : []),
  ];

  return (
    <div className="space-y-4">
      <Tabs tabs={subTabs} active={activeSub} onChange={setActiveSub} />
      {activeSub === "heroes" && <UnitGrid items={player.heroes ?? []} type="hero" />}
      {activeSub === "equipment" && player.heroEquipment && <UnitGrid items={player.heroEquipment} type="hero" />}
      {activeSub === "troops" && player.troops && <TroopSpellGrid items={player.troops} />}
      {activeSub === "spells" && player.spells && <TroopSpellGrid items={player.spells} />}
    </div>
  );
}

function UnitGrid({ items, type }: { items: HeroLevel[]; type: "hero" | "equipment" }) {
  if (!items.length) return <p className="text-text-muted text-sm p-4">暂无数据</p>;
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
      {items.map((h) => (
        <div key={h.name} className="flex items-center gap-3 p-3 bg-surface-page rounded-xl">
          <div className="w-10 h-10 rounded-xl bg-primary-bg flex items-center justify-center text-primary font-bold text-xs shrink-0 overflow-hidden">
            {h.name.slice(0, 2)}
          </div>
          <div className="min-w-0">
            <div className="text-sm font-medium truncate">{h.name}</div>
            <div className="text-xs text-text-secondary tabular-nums">
              <span className="text-primary font-medium">{h.level}</span> / {h.maxLevel}
            </div>
            {type !== "equipment" && h.village && (
              <div className="text-[10px] text-text-muted">{h.village === "home" ? "家园" : "夜世界"}</div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

function TroopSpellGrid({ items }: { items: TroopSpellLevel[] }) {
  const homeItems = items.filter((t) => t.village !== "builderBase");
  const bbItems = items.filter((t) => t.village === "builderBase");

  return (
    <div className="space-y-6">
      {homeItems.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-text-secondary mb-3">家园</h4>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {homeItems.map((t) => (
              <TroopCard key={t.name} item={t} />
            ))}
          </div>
        </div>
      )}
      {bbItems.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-text-secondary mb-3">夜世界</h4>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {bbItems.map((t) => (
              <TroopCard key={t.name} item={t} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function TroopCard({ item }: { item: TroopSpellLevel }) {
  return (
    <div className="p-3 bg-surface-page rounded-xl">
      <div className="flex items-center gap-2 mb-1">
        {item.superTroopActive && <Badge color="purple">超</Badge>}
        <span className="text-sm font-medium text-text-primary truncate">{item.name}</span>
      </div>
      <div className="text-xs text-text-secondary tabular-nums flex justify-between">
        <span>
          <span className="text-primary font-medium">{item.level}</span> / {item.maxLevel}
        </span>
      </div>
      {item.equipment && item.equipment.length > 0 && (
        <div className="mt-1.5 pt-1.5 border-t border-border space-y-0.5">
          {item.equipment.map((eq) => (
            <div key={eq.name} className="flex justify-between text-[11px] text-text-muted tabular-nums">
              <span className="truncate">{eq.name}</span>
              <span>{eq.level}/{eq.maxLevel}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function AchievementsTab({ achievements }: { achievements: AchievementProgress[] }) {
  if (!achievements.length) return <p className="text-text-muted text-sm p-4">暂无成就</p>;

  return (
    <div className="space-y-2">
      {achievements.map((a) => (
        <div
          key={a.name}
          className={`flex items-center gap-3 p-3 rounded-lg text-sm ${
            a.stars > 0 ? "bg-surface-page" : "bg-surface-page opacity-40"
          }`}
        >
          <span className="shrink-0 text-yellow-500">{"⭐".repeat(Math.min(a.stars, 3))}</span>
          <div className="flex-1 min-w-0">
            <div className="font-medium text-text-primary truncate">{a.name}</div>
            {a.info && <div className="text-xs text-text-muted truncate">{a.info}</div>}
          </div>
          <div className="text-xs text-text-secondary tabular-nums shrink-0">
            {a.value}/{a.target}
          </div>
          {a.village && (
            <span className="text-[10px] text-text-muted bg-border rounded px-1.5 py-0.5 shrink-0">{a.village === "home" ? "家园" : "夜世界"}</span>
          )}
        </div>
      ))}
    </div>
  );
}

function BattleLogSection({ tag }: { tag: string }) {
  const { data, isLoading } = useBattleLog(tag);

  return (
    <Card>
      <Card.Header><h3 className="font-semibold text-text-primary">战斗日志</h3></Card.Header>
      <Table
        loading={isLoading}
        emptyText="暂无战斗记录"
        columns={[
          { key: "battleTime", header: "时间", width: "120px", render: (row: BattleLogEntry) => {
            const ts = typeof row.battleTime === "number" ? String(row.battleTime) : row.battleTime;
            return <span className="text-xs text-text-secondary">{formatDateTime(ts)}</span>;
          }},
          { key: "type", header: "类型", width: "52px", render: (row: BattleLogEntry) => (
            <Badge color={row.attack ? "green" : "red"}>{row.attack ? "进攻" : "防守"}</Badge>
          )},
          { key: "stars", header: "星", width: "64px", render: (row: BattleLogEntry) => "⭐".repeat(row.stars) },
          { key: "destruction", header: "摧毁", width: "56px", render: (row: BattleLogEntry) => `${row.destructionPercentage}%` },
          { key: "opponent", header: "对手", width: "1fr", render: (row: BattleLogEntry) => (
            <span className="text-left block truncate font-medium">{row.opponentName}</span>
          )},
          { key: "opponentTH", header: "TH", width: "44px", render: (row: BattleLogEntry) => String(row.opponentTownHallLevel) },
        ]}
        data={data?.items ?? []}
      />
    </Card>
  );
}

function LeagueGroupSection({ tag }: { tag: string }) {
  const { data, isLoading } = useLeagueGroup(tag);

  if (isLoading) return (
    <Card>
      <Card.Header><h3 className="font-semibold text-text-primary">联赛分组</h3></Card.Header>
      <Card.Body><Skeleton.Table /></Card.Body>
    </Card>
  );
  if (!data?.members?.length) return (
    <Card>
      <Card.Header><h3 className="font-semibold text-text-primary">联赛分组</h3></Card.Header>
      <Card.Body><p className="text-text-muted text-sm">暂无联赛数据</p></Card.Body>
    </Card>
  );

  return (
    <Card>
      <Card.Header><h3 className="font-semibold text-text-primary">联赛分组</h3></Card.Header>
      <Table
        columns={[
          { key: "player", header: "玩家", width: "1fr", render: (row: LeagueGroupMember) => (
            <span className="text-left block truncate font-medium">{row.playerName}</span>
          )},
          { key: "clan", header: "部落", width: "120px", render: (row: LeagueGroupMember) => (
            <span className="text-text-secondary text-xs">{row.clanName}</span>
          )},
          { key: "trophies", header: "奖杯", width: "72px", render: (row: LeagueGroupMember) => formatNumber(row.leagueTrophies) },
          { key: "attack", header: "进攻", width: "90px", render: (row: LeagueGroupMember) => `${row.attackWinCount}W ${row.attackLoseCount}L` },
          { key: "defense", header: "防守", width: "90px", render: (row: LeagueGroupMember) => `${row.defenseWinCount}W ${row.defenseLoseCount}L` },
        ]}
        data={data.members}
      />
    </Card>
  );
}
