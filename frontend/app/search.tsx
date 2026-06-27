import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useSearchClans } from "~/lib/queries";
import { Card } from "~/components/ui/Card";
import { Button } from "~/components/ui/Button";
import { Table } from "~/components/ui/Table";
import { Badge } from "~/components/ui/Badge";
import { ErrorCard } from "~/components/ui/ErrorCard";
import { Skeleton } from "~/components/ui/Skeleton";
import { formatNumber } from "~/lib/utils";
import type { ClanOverview, ClanSearchParams } from "~/lib/types";

const WAR_FREQ_OPTIONS = [
  { value: "", label: "不限" },
  { value: "always", label: "一直打仗" },
  { value: "moreThanOncePerWeek", label: "每周多次" },
  { value: "oncePerWeek", label: "每周一次" },
  { value: "lessThanOncePerWeek", label: "少于每周一次" },
  { value: "never", label: "从不" },
];

export default function SearchPage() {
  const navigate = useNavigate();
  const [searchType, setSearchType] = useState<"clan" | "player">("clan");

  const [name, setName] = useState("");
  const [warFreq, setWarFreq] = useState("");
  const [minM, setMinM] = useState("");
  const [minLv, setMinLv] = useState("");
  const [minPts, setMinPts] = useState("");

  const [submitted, setSubmitted] = useState(false);
  const [queryParams, setQueryParams] = useState<ClanSearchParams>({});

  const { data, isLoading, isError, refetch } = useSearchClans(queryParams);

  const doSearch = () => {
    if (searchType === "player") {
      const tag = name.trim().toUpperCase();
      if (!tag) return;
      const clean = tag.startsWith("#") ? tag : `#${tag}`;
      navigate({ to: "/players/$tag", params: { tag: clean } });
      return;
    }
    const params: ClanSearchParams = {
      name: name || undefined,
      warFrequency: warFreq || undefined,
      minMembers: minM ? Number(minM) : undefined,
      minClanLevel: minLv ? Number(minLv) : undefined,
      minClanPoints: minPts ? Number(minPts) : undefined,
    };
    setQueryParams(params);
    setSubmitted(true);
  };

  const handleReset = () => {
    setName("");
    setWarFreq("");
    setMinM("");
    setMinLv("");
    setMinPts("");
    setSubmitted(false);
    setQueryParams({});
    navigate({ to: "/search" });
  };

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <div className="flex items-center gap-4">
        <h1 className="text-2xl font-bold text-text-primary">
          {searchType === "clan" ? "部落搜索" : "玩家查找"}
        </h1>
        <div className="flex rounded-btn border border-border overflow-hidden">
          <button
            type="button"
            onClick={() => { setSearchType("clan"); setSubmitted(false); }}
            className={`px-4 py-1.5 text-sm font-medium transition-colors ${searchType === "clan" ? "bg-primary text-white" : "text-text-secondary hover:text-text-primary"}`}
          >
            部落
          </button>
          <button
            type="button"
            onClick={() => { setSearchType("player"); setSubmitted(false); }}
            className={`px-4 py-1.5 text-sm font-medium transition-colors ${searchType === "player" ? "bg-primary text-white" : "text-text-secondary hover:text-text-primary"}`}
          >
            玩家
          </button>
        </div>
      </div>

      <Card>
        <Card.Body>
          {searchType === "player" ? (
            <div className="flex gap-3">
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") doSearch(); }}
                placeholder="输入玩家标签 #8GR..."
                className="flex-1 h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
              <Button onClick={doSearch}>查找</Button>
            </div>
          ) : (
            <div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-text-secondary font-medium">部落名称</label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    onKeyDown={(e) => { if (e.key === "Enter") doSearch(); }}
                    placeholder="搜索部落..."
                    className="h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-text-secondary font-medium">战争频率</label>
                  <select
                    value={warFreq}
                    onChange={(e) => setWarFreq(e.target.value)}
                    className="h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30"
                  >
                    {WAR_FREQ_OPTIONS.map((o) => (
                      <option key={o.value} value={o.value}>{o.label}</option>
                    ))}
                  </select>
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-text-secondary font-medium">最少成员</label>
                  <input type="number" value={minM} onChange={(e) => setMinM(e.target.value)} className="h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30" />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-text-secondary font-medium">最少等级</label>
                  <input type="number" value={minLv} onChange={(e) => setMinLv(e.target.value)} className="h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30" />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs text-text-secondary font-medium">最少积分</label>
                  <input type="number" value={minPts} onChange={(e) => setMinPts(e.target.value)} className="h-10 px-3 text-sm border border-border rounded-btn bg-white focus:outline-none focus:ring-2 focus:ring-primary/30" />
                </div>
              </div>
              <div className="flex gap-3 mt-4">
                <Button onClick={doSearch}>搜索</Button>
                <Button variant="ghost" onClick={handleReset}>重置</Button>
              </div>
            </div>
          )}
        </Card.Body>
      </Card>

      {searchType === "player" ? null : (
        <>
          {isError && <ErrorCard onRetry={() => refetch()} />}
          {isLoading && <Skeleton.Table rows={5} cols={4} />}

          {submitted && !isLoading && !isError && (data?.items?.length ? (
            <Card>
              <Card.Header><span className="text-text-secondary text-sm">共 {data.items.length} 个结果</span></Card.Header>
              <Table
                columns={[
                  { key: "badge", header: "", width: "48px",
                    render: (row: ClanOverview) => <img src={row.badgeUrls.small} alt="" className="w-8 h-8 rounded object-contain" /> },
                  { key: "name", header: "部落", width: "1fr", render: (row: ClanOverview) => (
                    <div className="text-left">
                      <a href={`/clans/${row.tag.replace("#", "%23")}`} className="font-medium text-sm text-primary hover:underline">{row.name}</a>
                      <div className="text-xs text-text-secondary">{row.tag}</div>
                    </div>
                  )},
                  { key: "clanLevel", header: "等级", width: "52px",
                    render: (row: ClanOverview) => <Badge color="primary">Lv.{row.clanLevel}</Badge> },
                  { key: "members", header: "成员", width: "48px", render: (row: ClanOverview) => `${row.members}` },
                  { key: "clanPoints", header: "积分", width: "72px", render: (row: ClanOverview) => (
                    <span className="tabular-nums">{formatNumber(row.clanPoints)}</span>
                  )},
                ]}
                data={data.items}
              />
            </Card>
          ) : (
            <div className="text-center text-text-muted text-sm py-10">未找到匹配的部落</div>
          ))}
        </>
      )}
    </div>
  );
}
