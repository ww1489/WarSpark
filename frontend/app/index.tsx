import { Link, useNavigate } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { Button } from "~/components/ui/Button";
import { Card } from "~/components/ui/Card";
import { Skeleton } from "~/components/ui/Skeleton";
import { BatchAvatarSelect } from "~/components/ui/BatchAvatarSelect";
import { useLocations, useHomeClanRankings, useHomePlayerRankings, useHomeCapitalRankings } from "~/lib/queries";
import { formatNumber } from "~/lib/utils";
import type { ClanRankingEntry, PlayerRankingEntry, ClanCapitalEntry } from "~/lib/types";

const DEFAULT_LOCATION = "32000017";

export default function HomePage() {
  return (
    <div>
      <HeroSection />
      <TopRankings />
    </div>
  );
}

function HeroSection() {
  const [value, setValue] = useState("");
  const [searchType, setSearchType] = useState<"clan" | "player">("clan");
  const navigate = useNavigate();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    const tag = value.trim().toUpperCase();
    if (!tag) return;
    const clean = tag.startsWith("#") ? tag : `#${tag}`;
    if (clean.length >= 4) {
      navigate({ to: searchType === "clan" ? "/clans/$tag" : "/players/$tag", params: { tag: clean } });
    }
    setValue("");
  };

  return (
    <div className="bg-gradient-to-br from-primary via-primary to-primary-light py-20 px-4">
      <div className="max-w-page mx-auto text-center">
        <h1 className="text-4xl font-bold text-white mb-4">WarSpark</h1>
        <p className="text-white/80 text-lg mb-8">Clash of Clans 数据查询平台</p>
        <form onSubmit={handleSearch} className="max-w-md mx-auto">
          <div className="flex justify-center gap-4 mb-4">
            <label className="flex items-center gap-2 text-white text-sm cursor-pointer">
              <input
                type="radio"
                name="searchType"
                checked={searchType === "clan"}
                onChange={() => setSearchType("clan")}
                className="accent-white"
              />
              部落
            </label>
            <label className="flex items-center gap-2 text-white text-sm cursor-pointer">
              <input
                type="radio"
                name="searchType"
                checked={searchType === "player"}
                onChange={() => setSearchType("player")}
                className="accent-white"
              />
              玩家
            </label>
          </div>
          <div className="flex gap-3">
            <input
              type="text"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={searchType === "clan" ? "输入部落标签 #2PP..." : "输入玩家标签 #8GR..."}
              className="flex-1 h-12 px-6 bg-white text-text-primary rounded-btn border-none text-sm focus:outline-none focus:ring-2 focus:ring-white/30"
            />
            <Button type="submit" size="lg" variant="secondary" className="border-white text-white hover:bg-white/10 bg-white/10">
              搜索
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

function TopRankings() {
  const [locationId, setLocationId] = useState(DEFAULT_LOCATION);
  const { data: locData } = useLocations();

  const locationOptions = useMemo(() => {
    if (!locData?.items) return [];
    return locData.items.map((l) => ({ value: String(l.id), label: l.name }));
  }, [locData]);

  return (
    <div className="max-w-page mx-auto py-12 px-4">
      <div className="flex items-center gap-4 mb-8">
        <h2 className="text-2xl font-bold text-text-primary">热门排行榜</h2>
        <BatchAvatarSelect value={locationId} onChange={setLocationId} options={locationOptions} />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <ClanRankingColumn locationId={locationId} />
        <PlayerRankingColumn locationId={locationId} />
        <CapitalRankingColumn locationId={locationId} />
      </div>
    </div>
  );
}

function RankBadge({ i, rank }: { i: number; rank: number }) {
  const colors = ["bg-yellow-500", "bg-gray-400", "bg-amber-600"];
  return (
    <span className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold text-white ${colors[i] ?? "bg-gray-300 text-gray-500"}`}>
      {rank}
    </span>
  );
}

function ClanRankingColumn({ locationId }: { locationId: string }) {
  const { data, isLoading } = useHomeClanRankings(locationId);

  return (
    <Link to="/leaderboards" className="block">
      <Card className="card-hover h-full">
        <Card.Header><h3 className="font-semibold text-text-primary">TOP 部落</h3></Card.Header>
        <Card.Body>
          {isLoading && <Skeleton className="h-60 w-full" />}
          {!data?.items.length ? <div className="py-10 text-center text-text-secondary text-sm">暂无数据</div> : (
            <div className="space-y-2">
              {(data.items as ClanRankingEntry[]).slice(0, 10).map((item, i) => (
                <div key={item.tag} className="flex items-center gap-3 text-sm py-1">
                  <RankBadge i={i} rank={item.rank} />
                  <img src={item.badgeUrls.small} alt="" className="w-6 h-6 rounded object-contain" />
                  <span className="flex-1 truncate font-medium text-text-primary">{item.name}</span>
                  <span className="tabular-nums text-text-secondary text-xs">{formatNumber(item.clanPoints)} 积分</span>
                </div>
              ))}
            </div>
          )}
        </Card.Body>
        <Card.Footer><span className="text-sm text-primary">查看更多 →</span></Card.Footer>
      </Card>
    </Link>
  );
}

function PlayerRankingColumn({ locationId }: { locationId: string }) {
  const { data, isLoading } = useHomePlayerRankings(locationId);

  return (
    <Link to="/leaderboards?type=players" className="block">
      <Card className="card-hover h-full">
        <Card.Header><h3 className="font-semibold text-text-primary">TOP 玩家</h3></Card.Header>
        <Card.Body>
          {isLoading && <Skeleton className="h-60 w-full" />}
          {!data?.items.length ? <div className="py-10 text-center text-text-secondary text-sm">暂无数据</div> : (
            <div className="space-y-2">
              {(data.items as PlayerRankingEntry[]).slice(0, 10).map((item, i) => (
                <div key={item.tag} className="flex items-center gap-2 text-sm py-1">
                  <RankBadge i={i} rank={item.rank} />
                  {item.leagueTier?.iconUrls?.small ? (
                    <img src={item.leagueTier.iconUrls.small} alt="" className="w-4 h-4 object-contain shrink-0" />
                  ) : null}
                  <span className="flex-1 truncate font-medium text-text-primary">{item.name}</span>
                  <span className="tabular-nums text-text-secondary text-xs">{formatNumber(item.trophies)}</span>
                </div>
              ))}
            </div>
          )}
        </Card.Body>
        <Card.Footer><span className="text-sm text-primary">查看更多 →</span></Card.Footer>
      </Card>
    </Link>
  );
}

function CapitalRankingColumn({ locationId }: { locationId: string }) {
  const { data, isLoading } = useHomeCapitalRankings(locationId);

  return (
    <Link to="/leaderboards?type=capital" className="block">
      <Card className="card-hover h-full">
        <Card.Header><h3 className="font-semibold text-text-primary">TOP 都城</h3></Card.Header>
        <Card.Body>
          {isLoading && <Skeleton className="h-60 w-full" />}
          {!data?.items.length ? <div className="py-10 text-center text-text-secondary text-sm">暂无数据</div> : (
            <div className="space-y-2">
              {(data.items as ClanCapitalEntry[]).slice(0, 10).map((item, i) => (
                <div key={item.tag} className="flex items-center gap-3 text-sm py-1">
                  <RankBadge i={i} rank={item.rank} />
                  <img src={item.badgeUrls.small} alt="" className="w-6 h-6 rounded object-contain" />
                  <span className="flex-1 truncate font-medium text-text-primary">{item.name}</span>
                  <span className="tabular-nums text-text-secondary text-xs">{formatNumber(item.clanCapitalPoints)} CP</span>
                </div>
              ))}
            </div>
          )}
        </Card.Body>
        <Card.Footer><span className="text-sm text-primary">查看更多 →</span></Card.Footer>
      </Card>
    </Link>
  );
}
